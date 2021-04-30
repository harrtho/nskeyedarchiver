package nskeyedarchiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	plist "github.com/xDevThomas/go-plist"
)

// toUIDList type asserts a []interface{} to a []plist.UID by iterating through the list.
func toUIDList(list []interface{}) []plist.UID {
	l := len(list)
	result := make([]plist.UID, l)
	for i := 0; i < l; i++ {
		result[i] = list[i].(plist.UID)
	}
	return result
}

// plistFromBytes decodes a binary or XML based PLIST and returns an interface{} or propagates the error raised by the library.
func plistFromBytes(plistBytes []byte) (interface{}, error) {
	var test interface{}
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))

	err := decoder.Decode(&test)
	if err != nil {
		return test, err
	}
	return test, nil
}

// ToPlist returns a string containing the plist.
func ToPlist(data interface{}) string {
	buf := &bytes.Buffer{}
	encoder := plist.NewEncoder(buf)
	encoder.Encode(data)
	return buf.String()
}

// Print an object as JSON for debugging purposes, careful log.Fatal on error
func printAsJSON(obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Fatalf("Error while marshalling Json:%s", err)
	}
	fmt.Print(string(b))
}

// NSDateToTime takes an Apple Cocoa Core Data timestamp and returns a golang time
// https://www.epochconverter.com/coredata
func NSDateToTime(timestamp float64) (time.Time, error) {
	floatParts := strings.Split(fmt.Sprintf("%.3f", timestamp), ".")

	// Timestamp
	timeint, err := strconv.ParseInt(floatParts[0], 10, 64)
	if err != nil {
		return time.Now(), err
	}
	seconds := int64(timeint + 978310800)

	// Timestamp nanoseconds
	timenano, err := strconv.ParseInt(floatParts[1], 10, 64)
	if err != nil {
		return time.Now(), err
	}
	nanoseconds := timenano

	return time.Unix(seconds, nanoseconds), nil
}

// verifyCorrectArchiver makes sure the nsKeyedArchived plist has all the right keys and values and returns an error otherwise
func verifyCorrectArchiver(nsKeyedArchiverData map[string]interface{}) error {
	if val, ok := nsKeyedArchiverData["$archiver"]; !ok {
		return fmt.Errorf("invalid NSKeyedArchiverObject, missing key '%s'", "$archiver")
	} else if stringValue := val.(string); stringValue != "NSKeyedArchiver" {
		return fmt.Errorf("invalid value: %s for key '%s', expected: '%s'", stringValue, "$archiver", "NSKeyedArchiver")
	}

	if _, ok := nsKeyedArchiverData["$top"]; !ok {
		return fmt.Errorf("invalid NSKeyedArchiverObject, missing key '%s'", "$top")
	}

	if _, ok := nsKeyedArchiverData["$objects"]; !ok {
		return fmt.Errorf("invalid NSKeyedArchiverObject, missing key '%s'", "$objects")
	}

	if val, ok := nsKeyedArchiverData["$version"]; !ok {
		return fmt.Errorf("invalid NSKeyedArchiverObject, missing key '%s'", "$version")
	} else if stringValue := val.(uint64); stringValue != 100000 {
		return fmt.Errorf("invalid value: %d for key '%s', expected: '%d'", stringValue, "$version", 100000)
	}

	return nil
}
