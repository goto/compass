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

	result := Merge(dst, src, nil)

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
	arrayMergeConfig := map[string]string{
		"data.columns": "name",
	}

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

	result := Merge(dst, src, arrayMergeConfig)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data field is not a map")
	}

	columns, ok := data["columns"].([]interface{})
	if !ok {
		t.Fatal("columns field is not an array")
	}

	// Result should have 2 columns (from src), not 3
	if len(columns) != 2 {
		t.Errorf("Expected 2 columns (from src), got %d", len(columns))
	}

	calcColumn := columns[0].(map[string]interface{})
	if calcColumn["name"] != "calculation_id" {
		t.Errorf("Expected first column to be calculation_id, got %s", calcColumn["name"])
	}
	if calcColumn["description"] != "updated description" {
		t.Errorf("Expected description to be updated, got %s", calcColumn["description"])
	}
	// data_type should be preserved from dst during merge
	if calcColumn["data_type"] != "STRING" {
		t.Errorf("Expected data_type to be preserved, got %s", calcColumn["data_type"])
	}

	signalColumn := columns[1].(map[string]interface{})
	if signalColumn["name"] != "signal_name" {
		t.Errorf("Expected second column to be signal_name, got %s", signalColumn["name"])
	}
	if signalColumn["description"] != "new column" {
		t.Errorf("Expected description to be 'new column', got %s", signalColumn["description"])
	}

	// entity_type column should NOT be present as it's not in src
}

func TestMerge_ColumnsArrayPreservesOrder(t *testing.T) {
	arrayMergeConfig := map[string]string{
		"data.columns": "name",
	}

	dst := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col1", "type": "string"},
				map[string]interface{}{"name": "col2", "type": "int"},
				map[string]interface{}{"name": "col3", "type": "bool"},
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

	result := Merge(dst, src, arrayMergeConfig)

	columns := result["data"].(map[string]interface{})["columns"].([]interface{})

	// Since src only has 1 column (col2), result should only have 1 column
	if len(columns) != 1 {
		t.Errorf("Expected 1 column (only col2 from src), got %d", len(columns))
	}

	col2 := columns[0].(map[string]interface{})
	if col2["name"] != "col2" {
		t.Errorf("Expected col2, got %s", col2["name"])
	}

	// The updated field from src should be present
	if updated, ok := col2["updated"].(bool); !ok || !updated {
		t.Error("Expected col2 to have updated=true from src")
	}

	// The type field from dst should be preserved through merge
	if col2["type"] != "int" {
		t.Errorf("Expected col2 type to be preserved as 'int', got %v", col2["type"])
	}
}

func TestMerge_NonDataColumnsArrayReplacedNormally(t *testing.T) {
	arrayMergeConfig := map[string]string{
		"data.columns": "name",
	}

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

	result := Merge(dst, src, arrayMergeConfig)

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

	result := Merge(dst, src, nil)

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
	arrayMergeConfig := map[string]string{
		"data.columns": "name",
	}

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

	result := Merge(dst, src, arrayMergeConfig)

	data := result["data"].(map[string]interface{})
	columns := data["columns"].([]interface{})

	if len(columns) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(columns))
	}

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

func TestMerge_ColumnArrayReplace_FewerColumns(t *testing.T) {
	arrayMergeConfig := map[string]string{
		"data.columns": "name",
	}

	// dst has 4 columns
	dst := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col1", "type": "string"},
				map[string]interface{}{"name": "col2", "type": "int", "nullable": true},
				map[string]interface{}{"name": "col3", "type": "bool"},
				map[string]interface{}{"name": "col4", "type": "float"},
			},
		},
	}

	// src has only 1 column (col2)
	src := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col2", "description": "updated column"},
			},
		},
	}

	result := Merge(dst, src, arrayMergeConfig)

	columns := result["data"].(map[string]interface{})["columns"].([]interface{})

	// Result should only have 1 column (col2 from src)
	if len(columns) != 1 {
		t.Errorf("Expected 1 column (only col2), got %d", len(columns))
	}

	col := columns[0].(map[string]interface{})
	if col["name"] != "col2" {
		t.Errorf("Expected col2, got %s", col["name"])
	}

	// Should have the new description from src
	if col["description"] != "updated column" {
		t.Errorf("Expected description from src, got %v", col["description"])
	}

	// Should preserve existing fields from dst
	if col["type"] != "int" {
		t.Errorf("Expected type preserved from dst, got %v", col["type"])
	}
	if nullable, ok := col["nullable"].(bool); !ok || !nullable {
		t.Error("Expected nullable field preserved from dst")
	}
}

func TestMerge_ColumnArrayReplace_MoreColumns(t *testing.T) {
	arrayMergeConfig := map[string]string{
		"data.columns": "name",
	}

	// dst has 2 columns
	dst := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col1", "type": "string"},
				map[string]interface{}{"name": "col2", "type": "int"},
			},
		},
	}

	// src has 4 columns
	src := map[string]interface{}{
		"data": map[string]interface{}{
			"columns": []interface{}{
				map[string]interface{}{"name": "col1", "description": "updated"},
				map[string]interface{}{"name": "col2", "description": "also updated"},
				map[string]interface{}{"name": "col3", "type": "bool", "description": "new column"},
				map[string]interface{}{"name": "col4", "type": "float", "description": "another new column"},
			},
		},
	}

	result := Merge(dst, src, arrayMergeConfig)

	columns := result["data"].(map[string]interface{})["columns"].([]interface{})

	// Result should have 4 columns (all from src)
	if len(columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(columns))
	}

	// Check col1 - should be merged
	col1 := columns[0].(map[string]interface{})
	if col1["name"] != "col1" {
		t.Errorf("Expected col1, got %s", col1["name"])
	}
	if col1["description"] != "updated" {
		t.Error("Expected description from src")
	}
	if col1["type"] != "string" {
		t.Error("Expected type preserved from dst")
	}

	// Check col3 - should be new
	col3 := columns[2].(map[string]interface{})
	if col3["name"] != "col3" {
		t.Errorf("Expected col3, got %s", col3["name"])
	}
	if col3["description"] != "new column" {
		t.Error("Expected description from src")
	}
}

func TestMerge_CustomArrayMergeConfig(t *testing.T) {
	arrayMergeConfig := map[string]string{
		"data.columns": "name",
		"owners":       "email",
	}

	dst := map[string]interface{}{
		"owners": []interface{}{
			map[string]interface{}{
				"email": "alice@example.com",
				"role":  "admin",
			},
			map[string]interface{}{
				"email": "bob@example.com",
				"role":  "viewer",
			},
		},
	}

	src := map[string]interface{}{
		"owners": []interface{}{
			map[string]interface{}{
				"email":  "bob@example.com",
				"role":   "editor",
				"active": true,
			},
			map[string]interface{}{
				"email": "charlie@example.com",
				"role":  "viewer",
			},
		},
	}

	result := Merge(dst, src, arrayMergeConfig)

	owners := result["owners"].([]interface{})

	// Result should have 2 owners (from src), not 3
	if len(owners) != 2 {
		t.Fatalf("Expected 2 owners (from src), got %d", len(owners))
	}

	bob := owners[0].(map[string]interface{})
	if bob["email"] != "bob@example.com" || bob["role"] != "editor" {
		t.Error("Bob should be updated to editor")
	}
	if active, ok := bob["active"].(bool); !ok || !active {
		t.Error("Bob should have active=true")
	}

	charlie := owners[1].(map[string]interface{})
	if charlie["email"] != "charlie@example.com" || charlie["role"] != "viewer" {
		t.Error("Charlie should be added")
	}

	// Alice should NOT be present as she's not in src
}
