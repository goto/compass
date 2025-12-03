package mergemap

import (
	"reflect"
	"strings"
)

const maxDepth = 32

// Merge recursively merges the src and dst maps. Key conflicts are resolved by
// preferring src, or recursively descending, if both src and dst are maps.
// Arrays specified in arrayMergeConfig are merged by their identifier field.
func Merge(dst, src map[string]interface{}, arrayMergeConfig map[string]string) map[string]interface{} {
	return mergeRecursive(dst, src, 0, []string{}, arrayMergeConfig)
}

func mergeRecursive(dst, src map[string]interface{}, depth int, path []string, arrayMergeConfig map[string]string) map[string]interface{} {
	if depth > maxDepth {
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
				srcVal = mergeRecursive(dstMap, srcMap, depth+1, currentPath, arrayMergeConfig)
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
	dstSlice, srcSlice := convertToSlices(dst, src)
	if dstSlice == nil || srcSlice == nil {
		return nil
	}

	dstMap, dstOrder := buildDestinationMap(dstSlice, identifierField)
	updateMapWithSource(dstMap, &dstOrder, srcSlice, identifierField, arrayMergeConfig)

	return rebuildArray(dstMap, dstOrder)
}

// convertToSlices validates and converts interfaces to slices
func convertToSlices(dst, src interface{}) ([]interface{}, []interface{}) {
	dstSlice, dstOk := dst.([]interface{})
	srcSlice, srcOk := src.([]interface{})

	if !dstOk || !srcOk {
		return nil, nil
	}

	return dstSlice, srcSlice
}

// buildDestinationMap creates a map and order slice from destination items
func buildDestinationMap(dstSlice []interface{}, identifierField string) (map[string]map[string]interface{}, []string) {
	dstMap := make(map[string]map[string]interface{})
	dstOrder := []string{}

	for _, item := range dstSlice {
		itemMap, identifier := extractItemData(item, identifierField)
		if itemMap == nil || identifier == "" {
			continue
		}
		dstMap[identifier] = itemMap
		dstOrder = append(dstOrder, identifier)
	}

	return dstMap, dstOrder
}

// updateMapWithSource merges or adds source items to the destination map
func updateMapWithSource(
	dstMap map[string]map[string]interface{},
	dstOrder *[]string,
	srcSlice []interface{},
	identifierField string,
	arrayMergeConfig map[string]string,
) {
	for _, item := range srcSlice {
		itemMap, identifier := extractItemData(item, identifierField)
		if itemMap == nil || identifier == "" {
			continue
		}

		if existingItem, exists := dstMap[identifier]; exists {
			dstMap[identifier] = mergeRecursive(existingItem, itemMap, 0, []string{}, arrayMergeConfig)
		} else {
			dstMap[identifier] = itemMap
			*dstOrder = append(*dstOrder, identifier)
		}
	}
}

// extractItemData extracts and validates item data and identifier
func extractItemData(item interface{}, identifierField string) (map[string]interface{}, string) {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return nil, ""
	}

	identifier, ok := itemMap[identifierField].(string)
	if !ok || identifier == "" {
		return nil, ""
	}

	return itemMap, identifier
}

// rebuildArray reconstructs the array from the map while maintaining order
func rebuildArray(dstMap map[string]map[string]interface{}, dstOrder []string) []interface{} {
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
