package translator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestQueryExprTranslator_ConvertToEsQuery(t *testing.T) {
	tests := []struct {
		name                string
		queryExprTranslator *QueryExprTranslator
		want                string
		wantErr             bool
	}{
		{
			name: "less than condition",
			queryExprTranslator: &QueryExprTranslator{
				QueryExpr: `updated_at < "2024-04-05 23:59:59"`,
			},
			want:    "test-json/lt-condition.json",
			wantErr: false,
		},
		{
			name: "in condition",
			queryExprTranslator: &QueryExprTranslator{
				QueryExpr: `service in ["test1","test2","test3"]`,
			},
			want:    "test-json/in-condition.json",
			wantErr: false,
		},
		{
			name: "equals or not in condition",
			queryExprTranslator: &QueryExprTranslator{
				QueryExpr: `name == "John" || service not in ["test1","test2","test3"]`,
			},
			want:    "test-json/equals-or-not-in-condition.json",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.queryExprTranslator.ConvertToEsQuery()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToEsQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !deepEqual(tt.want, tt.queryExprTranslator.EsQuery) {
				t.Errorf("ConvertToEsQuery() got = %v, want equal to json in file: %v", got, tt.want)
			}
		})
	}
}

func TestQueryExprTranslator_ConvertToSQL(t *testing.T) {
	tests := []struct {
		name                string
		queryExprTranslator *QueryExprTranslator
		want                string
		wantErr             bool
	}{
		{
			name: "less than condition",
			queryExprTranslator: &QueryExprTranslator{
				QueryExpr: `updated_at < "2024-04-05 23:59:59"`,
			},
			want:    `(updated_at < '2024-04-05 23:59:59')`,
			wantErr: false,
		},
		{
			name: "in condition",
			queryExprTranslator: &QueryExprTranslator{
				QueryExpr: `service in ["test1","test2","test3"]`,
			},
			want:    `(service IN ('test1', 'test2', 'test3'))`,
			wantErr: false,
		},
		{
			name: "equals or not in condition",
			queryExprTranslator: &QueryExprTranslator{
				QueryExpr: `name == "John" || service not in ["test1","test2","test3"]`,
			},
			want:    `((name = 'John') OR (service NOT IN ('test1', 'test2', 'test3')))`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.queryExprTranslator.ConvertToSQL()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertToSQL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func deepEqual(jsonFileName string, result map[string]interface{}) bool {
	// Step 1: Read the JSON file
	fileContent, err := ioutil.ReadFile(jsonFileName)
	if err != nil {
		fmt.Println("Error reading the file:", err)
		return false
	}

	// Step 2: Unmarshal the file content into a Go data structure
	var fileData map[string]interface{}
	err = json.Unmarshal(fileContent, &fileData)
	if err != nil {
		fmt.Println("Error unmarshalling the file content:", err)
		return false
	}

	// Step 4: Compare the two Go data structures
	if reflect.DeepEqual(fileData, result) {
		return true
	} else {
		return false
	}
}
