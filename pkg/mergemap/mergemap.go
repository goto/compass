package mergemap

import (
	"reflect"
	"strings"
)

var (
	MaxDepth = 32
)

// Merge recursively merges the src and dst maps. Key conflicts are resolved by
// preferring src, or recursively descending, if both src and dst are maps.
// Arrays specified in arrayMergeConfig are merged by their identifier field.
func Merge(dst, src map[string]interface{}, arrayMergeConfig map[string]string) map[string]interface{} {
	return merge(dst, src, 0, []string{}, arrayMergeConfig)
}

func merge(dst, src map[string]interface{}, depth int, path []string, arrayMergeConfig map[string]string) map[string]interface{} {
	if depth > MaxDepth {
		panic("too deep!")
	}
	for key, srcVal := range src {
		currentPath := append(path, key)

		if dstVal, ok := dst[key]; ok {
			// Check if this array should be merged by identifier
			if identifierField := getIdentifierField(currentPath, arrayMergeConfig); identifierField != "" {
				if merged := mergeArraysByIdentifier(dstVal, srcVal, identifierField, arrayMergeConfig); merged != nil {
					dst[key] = merged
					continue
				}
			}

			srcMap, srcMapOk := mapify(srcVal)
			dstMap, dstMapOk := mapify(dstVal)
			if srcMapOk && dstMapOk {
				srcVal = merge(dstMap, srcMap, depth+1, currentPath, arrayMergeConfig)
			}
		}
		dst[key] = srcVal
	}
	return dst
}

// getIdentifierField checks if the current path matches any configured array merge path
// and returns the identifier field to use for merging, or empty string if not configured
func getIdentifierField(path []string, arrayMergeConfig map[string]string) string {
	if len(path) == 0 || arrayMergeConfig == nil {
		return ""
	}

	// Build the dot-separated path
	pathStr := strings.Join(path, ".")

	// Check if this path is configured for merge-by-identifier
	if identifierField, exists := arrayMergeConfig[pathStr]; exists {
		return identifierField
	}

	return ""
}

// mergeArraysByIdentifier merges two arrays by matching items with the same identifier field
func mergeArraysByIdentifier(dst, src interface{}, identifierField string, arrayMergeConfig map[string]string) interface{} {
	dstSlice, dstOk := dst.([]interface{})
	srcSlice, srcOk := src.([]interface{})

	if !dstOk || !srcOk {
		return nil
	}

	// Build a map of dst items by identifier field
	dstMap := make(map[string]map[string]interface{})
	dstOrder := []string{}

	for _, item := range dstSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		identifier, ok := itemMap[identifierField].(string)
		if !ok || identifier == "" {
			continue
		}
		dstMap[identifier] = itemMap
		dstOrder = append(dstOrder, identifier)
	}

	// Merge or add src items
	for _, item := range srcSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		identifier, ok := itemMap[identifierField].(string)
		if !ok || identifier == "" {
			continue
		}

		if existingItem, exists := dstMap[identifier]; exists {
			// Merge the existing item with new data
			dstMap[identifier] = merge(existingItem, itemMap, 0, []string{}, arrayMergeConfig)
		} else {
			// Add new item
			dstMap[identifier] = itemMap
			dstOrder = append(dstOrder, identifier)
		}
	}

	// Rebuild the array maintaining order
	result := make([]interface{}, 0, len(dstMap))
	for _, identifier := range dstOrder {
		if item, exists := dstMap[identifier]; exists {
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
