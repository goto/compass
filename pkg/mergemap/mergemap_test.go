package mergemap

import (
	"reflect"
	"testing"
)

func TestMerge_BasicMerge(t *testing.T) {
	dst := map[string]interface{}{
		"name": "old",
		"age":  30,
	}
	src := map[string]interface{}{
		"age":  35,
		"city": "NYC",
	}

	result := Merge(dst, src)

	expected := map[string]interface{}{
		"name": "old",
		"age":  35,
		"city": "NYC",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMerge_ColumnsArrayByName(t *testing.T) {
	dst := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{
					"name":        "calculation_id",
					"data_type":   "STRING",
					"description": "old description",
					"is_nullable": true,
				},
				map[string]interface{}{
					"name":        "entity_type",
					"data_type":   "STRING",
					"description": "",
					"is_nullable": true,
				},
			},
		},
	}

	src := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{
					"name":        "calculation_id",
					"description": "updated description",
				},
				map[string]interface{}{
					"name":        "signal_name",
					"data_type":   "STRING",
					"description": "new column",
					"is_nullable": true,
				},
			},
		},
	}

	result := Merge(dst, src)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data field is not a map")
	}

	columns, ok := data["columns"].([]interface{})
	if !ok {
		t.Fatal("columns field is not an array")
	}

	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	calcColumn := columns[0].(map[string]interface{})
	if calcColumn["name"] != "calculation_id" {
		t.Errorf("Expected first column to be calculation_id, got %s", calcColumn["name"])
	}
	if calcColumn["description"] != "updated description" {
		t.Errorf("Expected description to be updated, got %s", calcColumn["description"])
	}
	if calcColumn["data_type"] != "STRING" {
		t.Errorf("Expected data_type to be preserved, got %s", calcColumn["data_type"])
	}

	entityColumn := columns[1].(map[string]interface{})
	if entityColumn["name"] != "entity_type" {
		t.Errorf("Expected second column to be entity_type, got %s", entityColumn["name"])
	}

	signalColumn := columns[2].(map[string]interface{})
	if signalColumn["name"] != "signal_name" {
		t.Errorf("Expected third column to be signal_name, got %s", signalColumn["name"])
	}
	if signalColumn["description"] != "new column" {
		t.Errorf("Expected description to be 'new column', got %s", signalColumn["description"])
	}
}

func TestMerge_ColumnsArrayPreservesOrder(t *testing.T) {
	dst := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col1"},
				map[string]interface{}{"name": "col2"},
				map[string]interface{}{"name": "col3"},
			},
		},
	}

	src := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col2", "updated": true},
			},
		},
	}

	result := Merge(dst, src)

	columns := result["data"].(map[string]interface{})["columns"].([]interface{})

	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	names := []string{}
	for _, col := range columns {
		names = append(names, col.(map[string]interface{})["name"].(string))
	}

	expectedNames := []string{"col1", "col2", "col3"}
	if !reflect.DeepEqual(names, expectedNames) {
		t.Errorf("Expected order %v, got %v", expectedNames, names)
	}

	col2 := columns[1].(map[string]interface{})
	if updated, ok := col2["updated"].(bool); !ok || !updated {
		t.Error("Expected col2 to be updated")
	}
}

func TestMerge_NonDataColumnsArrayReplacedNormally(t *testing.T) {
	dst := map[string]interface{}{
		"other": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col1"},
			},
		},
	}

	src := map[string]interface{}{
		"other": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col2"},
			},
		},
	}

	result := Merge(dst, src)

	columns := result["other"].(map[string]interface{})["columns"].([]interface{})

	if len(columns) != 1 {
		t.Errorf("Expected 1 column (replaced), got %d", len(columns))
	}

	col := columns[0].(map[string]interface{})
	if col["name"] != "col2" {
		t.Errorf("Expected col2 (replaced), got %s", col["name"])
	}
}

func TestMerge_NestedMapsStillWork(t *testing.T) {
	dst := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	src := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"key2": "updated",
				"key3": "value3",
			},
		},
	}

	result := Merge(dst, src)

	attrs := result["data"].(map[string]interface{})["attributes"].(map[string]interface{})

	if attrs["key1"] != "value1" {
		t.Error("key1 should be preserved")
	}
	if attrs["key2"] != "updated" {
		t.Error("key2 should be updated")
	}
	if attrs["key3"] != "value3" {
		t.Error("key3 should be added")
	}
}

func TestMerge_ColumnAttributesMerged(t *testing.T) {
	dst := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{
					"name":       "calculation_id",
					"data_type":  "STRING",
					"attributes": map[string]interface{}{},
				},
				map[string]interface{}{
					"name":      "entity_type",
					"data_type": "STRING",
					"attributes": map[string]interface{}{
						"masking_policy": []interface{}{"test", "test_2"},
					},
				},
				map[string]interface{}{
					"name":       "latest_status_name",
					"data_type":  "STRING",
					"attributes": map[string]interface{}{},
				},
			},
		},
	}

	src := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{
					"name":       "calculation_id",
					"data_type":  "STRING",
					"attributes": map[string]interface{}{},
				},
				map[string]interface{}{
					"name":      "entity_type",
					"data_type": "STRING",
					"attributes": map[string]interface{}{
						"masking_policy": []interface{}{"test", "test_2"},
						"masking_roles":  []interface{}{"test_role"},
					},
				},
				map[string]interface{}{
					"name":       "latest_status_name",
					"data_type":  "STRING",
					"attributes": map[string]interface{}{},
				},
			},
		},
	}

	result := Merge(dst, src)

	data := result["data"].(map[string]interface{})
	columns := data["columns"].([]interface{})

	// Check that we still have 3 columns
	if len(columns) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(columns))
	}

	// Find entity_type column
	var entityTypeCol map[string]interface{}
	for _, col := range columns {
		colMap := col.(map[string]interface{})
		if colMap["name"] == "entity_type" {
			entityTypeCol = colMap
			break
		}
	}

	if entityTypeCol == nil {
		t.Fatal("entity_type column not found")
	}

	attrs := entityTypeCol["attributes"].(map[string]interface{})

	if _, ok := attrs["masking_policy"]; !ok {
		t.Error("masking_policy should be present")
	}

	if _, ok := attrs["masking_roles"]; !ok {
		t.Error("masking_roles should be present after merge")
	}

	maskingRoles := attrs["masking_roles"].([]interface{})
	if len(maskingRoles) != 1 || maskingRoles[0] != "test_role" {
		t.Errorf("Expected masking_roles to be ['test_role'], got %v", maskingRoles)
	}

	maskingPolicy := attrs["masking_policy"].([]interface{})
	if len(maskingPolicy) != 2 {
		t.Errorf("Expected masking_policy to have 2 items, got %d", len(maskingPolicy))
	}
}
