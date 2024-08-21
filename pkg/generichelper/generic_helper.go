package generichelper

import (
	"reflect"
)

// Contains checks if a target item exists in an array of any type.
//
// Example
//
//	names := []string{"Alice", "Bob", "Carol"}
//	result := Contains(names, "Bob")
//
// Result:
//
//	true
func Contains[T comparable](arr []T, target T) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

// GetJSONTags retrieves all JSON tag values from a struct.
// It returns a slice of JSON tag values extracted from the struct's fields.
//
// Example:
//
//		type Person struct {
//	  ID        int    `json:"id"`
//	  Name      string `json:"name"`
//	  Age       int    `json:"age"`
//	  CreatedAt string `json:"created_at"`
//		}
//
//		p := Person{}
//	 jsonTags := GetJSONTags(p)
//
// Result:
//
//	["id", "name", "age", "created_at"]
func GetJSONTags(v interface{}) []string {
	var tags []string
	val := reflect.ValueOf(v)
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			tags = append(tags, jsonTag)
		}
	}

	return tags
}

// GetMapKeys is a generic function that extracts all keys from a map and returns them in a slice.
//
// Example:
//
//	ageMap := map[string]int{"Alice": 30, "Bob": 25, "Carol": 27}
//	keys := GetMapKeys(ageMap)
//
// Result:
//
//	["Alice", "Bob", "Carol"]
func GetMapKeys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
