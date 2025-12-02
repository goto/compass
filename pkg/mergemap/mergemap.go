package mergemap

import (
	"reflect"
)

var MaxDepth = 32

// Merge recursively merges the src and dst maps. Key conflicts are resolved by
// preferring src, or recursively descending, if both src and dst are maps.
// Special handling: data.columns arrays are merged by name field.
func Merge(dst, src map[string]interface{}) map[string]interface{} {
	return merge(dst, src, 0, []string{})
}

func merge(dst, src map[string]interface{}, depth int, path []string) map[string]interface{} {
	if depth > MaxDepth {
		panic("too deep!")
	}
	for key, srcVal := range src {
		currentPath := append(path, key)

		if dstVal, ok := dst[key]; ok {
			// Special handling for data.columns - merge by name
			if shouldMergeByName(currentPath, key) {
				if merged := mergeArraysByName(dstVal, srcVal); merged != nil {
					dst[key] = merged
					continue
				}
			}

			srcMap, srcMapOk := mapify(srcVal)
			dstMap, dstMapOk := mapify(dstVal)
			if srcMapOk && dstMapOk {
				srcVal = merge(dstMap, srcMap, depth+1, currentPath)
			}
		}
		dst[key] = srcVal
	}
	return dst
}

// shouldMergeByName checks if the current path matches data.columns
func shouldMergeByName(path []string, key string) bool {
	if key != "columns" {
		return false
	}
	// Check if this is data.columns (path would be ["data", "columns"])
	if len(path) >= 2 && path[len(path)-2] == "data" && path[len(path)-1] == "columns" {
		return true
	}
	return false
}

// mergeArraysByName merges two arrays by matching items with the same "name" field
func mergeArraysByName(dst, src interface{}) interface{} {
	dstSlice, dstOk := dst.([]interface{})
	srcSlice, srcOk := src.([]interface{})

	if !dstOk || !srcOk {
		return nil
	}

	// Build a map of dst items by name
	dstMap := make(map[string]map[string]interface{})
	dstOrder := []string{}

	for _, item := range dstSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := itemMap["name"].(string)
		if !ok || name == "" {
			continue
		}
		dstMap[name] = itemMap
		dstOrder = append(dstOrder, name)
	}

	// Merge or add src items
	for _, item := range srcSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := itemMap["name"].(string)
		if !ok || name == "" {
			continue
		}

		if existingItem, exists := dstMap[name]; exists {
			// Merge the existing item with new data
			dstMap[name] = merge(existingItem, itemMap, 0, []string{})
		} else {
			// Add new item
			dstMap[name] = itemMap
			dstOrder = append(dstOrder, name)
		}
	}

	// Rebuild the array maintaining order
	result := make([]interface{}, 0, len(dstMap))
	for _, name := range dstOrder {
		if item, exists := dstMap[name]; exists {
			result = append(result, item)
		}
	}

	return result
}

func mapify(i interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return map[string]interface{}{}, false
}
