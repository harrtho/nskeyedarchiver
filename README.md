# nskeyedarchiver

A fork extending [@danielpaulus's excellent Golang based implementation](https://github.com/danielpaulus/nskeyedarchiver) NSKeyedArchiver with support for extracting Time, String, and custom objects.

Unarchive extracts NSKeyedArchiver Plists, either in XML or Binary format, and returns an array of the archived objects converted to usable Go Types.
- Primitives will be extracted just like regular Plist primitives (string, float64, int64, []uint8 etc.).
- NSArray, NSMutableArray, NSSet and NSMutableSet will transformed into []interface{}
- NSDictionary and NSMutableDictionary will be transformed into map[string] interface{}. I might add non string keys later.

Todos: 
- Add custom object support (anything that is not an array, set or dictionary)
- Add archiving/encoding support

Unarchive example:
```golang
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/qcasey/nskeyedarchiver"
)

func main() {
	fileData, err := ioutil.ReadFile("apple-nskeyedariver.plist")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	plistData, err := nskeyedarchiver.Unarchive(fileData)
	if err != nil {
        fmt.Println("Error decoding plist:", err)
        return
    }
    
	for key, value := range plistData[0].(map[string]interface{}) {
		fmt.Printf("%s = %v (%T)\n", key, value, value)
	}
}
```


Thanks howett.net/plist for your awesome Plist library :-) 
