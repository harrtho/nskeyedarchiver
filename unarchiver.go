package nskeyedarchiver

import (
	"fmt"

	"github.com/rs/zerolog/log"
	plist "howett.net/plist"
)

//Unarchive extracts NSKeyedArchiver Plists, either in XML or Binary format, and returns an array of the archived objects converted to usable Go Types.
// Primitives will be extracted just like regular Plist primitives (string, float64, int64, []uint8 etc.).
// NSArray, NSMutableArray, NSSet and NSMutableSet will transformed into []interface{}
// NSDictionary and NSMutableDictionary will be transformed into map[string] interface{}. I might add non string keys later.
func Unarchive(xml []byte) ([]interface{}, error) {
	plist, err := plistFromBytes(xml)
	if err != nil {
		return nil, err
	}
	nsKeyedArchiverData := plist.(map[string]interface{})

	err = verifyCorrectArchiver(nsKeyedArchiverData)
	if err != nil {
		return nil, err
	}
	return extractObjectsFromTop(nsKeyedArchiverData["$top"].(map[string]interface{}), nsKeyedArchiverData["$objects"].([]interface{}))

}

func extractObjectsFromTop(top map[string]interface{}, objects []interface{}) ([]interface{}, error) {
	objectCount := len(top)
	if root, ok := top["root"]; ok {
		return extractObjects([]plist.UID{root.(plist.UID)}, objects)
	}
	objectRefs := make([]plist.UID, objectCount)
	//convert the Dictionary with the objectReferences into a flat list of UIDs, so we can reuse the extractObjects function later
	for i := 0; i < objectCount; i++ {
		objectIndex := top[fmt.Sprintf("$%d", i)].(plist.UID)
		objectRefs[i] = objectIndex
	}
	return extractObjects(objectRefs, objects)
}

func extractObjects(objectRefs []plist.UID, objects []interface{}) ([]interface{}, error) {
	objectCount := len(objectRefs)
	returnValue := make([]interface{}, objectCount)
	log.Debug().Msgf("Extracting %d objects from list of %d total objects\n", objectCount, len(objects))
	for i := 0; i < objectCount; i++ {
		objectIndex := objectRefs[i]
		objectRef := objects[objectIndex]

		if object, ok := isPrimitiveObject(objectRef); ok {
			returnValue[i] = object
			continue
		}

		objectInterface := objectRef.(map[string]interface{})
		if ok := isTimeObject(objectInterface, objects); ok {
			timestamp, err := nsDateToTime(objectInterface["NS.time"].(float64))
			if err != nil {
				return nil, err
			}
			returnValue[i] = timestamp
			continue
		}

		if ok := isStringObject(objectInterface, objects); ok {
			returnValue[i] = objectInterface["NS.string"].(string)
			continue
		}

		if object, ok := isArrayObject(objectInterface, objects); ok {
			extractObjects, err := extractObjects(toUIDList(object["NS.objects"].([]interface{})), objects)
			if err != nil {
				return nil, err
			}
			returnValue[i] = extractObjects
			continue
		}

		if ok := isDictionaryObject(objectInterface, objects); ok {
			dictionary, err := extractDictionary(objectInterface, objects)
			if err != nil {
				return nil, err
			}
			returnValue[i] = dictionary
			continue
		}

		customObject, err := extractCustomObject(objectInterface, objects)
		if err != nil {
			return nil, err
		}
		returnValue[i] = customObject
	}
	return returnValue, nil
}

func isArrayObject(object map[string]interface{}, objects []interface{}) (map[string]interface{}, bool) {
	className, err := resolveClass(object["$class"], objects)
	if err != nil {
		return nil, false
	}
	if className == "NSArray" || className == "NSMutableArray" || className == "NSSet" || className == "NSMutableSet" {
		return object, true
	}
	return object, false
}

func isDictionaryObject(object map[string]interface{}, objects []interface{}) bool {
	className, err := resolveClass(object["$class"], objects)
	if err != nil {
		return false
	}
	if className == "NSDictionary" || className == "NSMutableArray" || className == "NSMutableDictionary" {
		return true
	}
	return false
}

func isTimeObject(object map[string]interface{}, objects []interface{}) bool {
	className, err := resolveClass(object["$class"], objects)
	if err != nil {
		return false
	}
	if className == "NSDate" {
		return true
	}
	return false
}

// Support for NS.String
// 26 => {
// 	"$classes" => [
// 	  0 => "NSMutableString"
// 	  1 => "NSString"
// 	  2 => "NSObject"
// 	]
// 	"$classname" => "NSMutableString"
// }
func isStringObject(object map[string]interface{}, objects []interface{}) bool {
	className, err := resolveClass(object["$class"], objects)
	if err != nil {
		return false
	}
	if className == "NSMutableString" {
		return true
	}
	return false
}

func extractDictionary(object map[string]interface{}, objects []interface{}) (map[string]interface{}, error) {
	keyRefs := toUIDList(object["NS.keys"].([]interface{}))
	keys, err := extractObjects(keyRefs, objects)
	if err != nil {
		return nil, err
	}

	valueRefs := toUIDList(object["NS.objects"].([]interface{}))
	values, err := extractObjects(valueRefs, objects)
	if err != nil {
		return nil, err
	}
	mapSize := len(keys)
	result := make(map[string]interface{}, mapSize)
	for i := 0; i < mapSize; i++ {
		result[keys[i].(string)] = values[i]
	}

	return result, nil
}

// Custom Object, where the keys are the map indexes, i.e.
// "$class" => <CFKeyedArchiverUID 0x7f8383e07f60 [0x7fff8912ccc0]>{value = 56}
// "albumGUID" => <CFKeyedArchiverUID 0x7f8383e07ea0 [0x7fff8912ccc0]>{value = 4}
// "assets" => <CFKeyedArchiverUID 0x7f8383e07e40 [0x7fff8912ccc0]>{value = 5}
// "ctag" => <CFKeyedArchiverUID 0x7f8383e07ec0 [0x7fff8912ccc0]>{value = 3}
// "email" => <CFKeyedArchiverUID 0x7f8383e07fc0 [0x7fff8912ccc0]>{value = 48}
func extractCustomObject(object map[string]interface{}, objects []interface{}) (map[string]interface{}, error) {

	// Extract keys from their initial place, build array of UIDs to extract
	objectPrimitives := make(map[string]interface{}, len(object))
	var objectValueList []interface{}
	var objectKeyList []string
	for key, value := range object {
		if key == "$class" || key == "$classes" {
			log.Debug().Msgf("Ignoring class definition %v\n", key)
		} else if _, ok := isPrimitiveObject(value); ok {
			log.Debug().Msgf("Adding primitive directly %v:%v\n", key, value)
			objectPrimitives[key] = value
		} else {
			objectValueList = append(objectValueList, value)
			objectKeyList = append(objectKeyList, key)
		}
	}

	valueRefs := toUIDList(objectValueList)
	values, err := extractObjects(valueRefs, objects)
	if err != nil {
		return nil, err
	}
	mapSize := len(values)
	result := make(map[string]interface{}, len(objectPrimitives)+mapSize)

	// Add primitives
	for key, value := range objectPrimitives {
		result[key] = value
	}

	// Add values extracted from UIDs
	for i := 0; i < mapSize; i++ {
		result[objectKeyList[i]] = values[i]
	}

	return result, nil
}

func resolveClass(classInfo interface{}, objects []interface{}) (string, error) {
	if v, ok := classInfo.(plist.UID); ok {
		classDict := objects[v].(map[string]interface{})
		return classDict["$classname"].(string), nil
	}
	return "", fmt.Errorf("Could not find class for %s", classInfo)
}

func isPrimitiveObject(object interface{}) (interface{}, bool) {
	if v, ok := object.(uint64); ok {
		return v, ok
	}
	if v, ok := object.(float64); ok {
		return v, ok
	}
	if v, ok := object.(bool); ok {
		return v, ok
	}
	if v, ok := object.(string); ok {
		return v, ok
	}
	if v, ok := object.([]uint8); ok {
		return v, ok
	}
	return object, false
}
