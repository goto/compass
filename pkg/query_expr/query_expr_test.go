package queryexpr_test

import (
	"reflect"
	"testing"

	queryexpr "github.com/goto/compass/pkg/query_expr"
)

func TestGetIdentifiersMap(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "got 0 identifier test",
			expr:    `findLast([1, 2, 3, 4], # > 2)`,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "got 1 identifiers test",
			expr: `updated_at < '2024-04-05 23:59:59'`,
			want: map[string]string{
				"updated_at": "<",
			},
			wantErr: false,
		},
		{
			name: "got 3 identifiers test",
			expr: `(identifier1 == !(findLast([1, 2, 3, 4], # > 2) == 4)) && identifier2 != 'John' || identifier3 == "hallo"`,
			want: map[string]string{
				"identifier1": "==",
				"identifier2": "!=",
				"identifier3": "==",
			},
			wantErr: false,
		},
		{
			name:    "got error",
			expr:    `findLast([1, 2, 3, 4], # > 2`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queryexpr.GetIdentifiersMap(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIdentifiersMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIdentifiersMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetQueryExprResult(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    any
		wantErr bool
	}{
		{
			name:    "return a value from func",
			expr:    `findLast([1, 2, 3, 4], # > 2)`,
			want:    4,
			wantErr: false,
		},
		{
			name:    "return a value func equation",
			expr:    `false == !true`,
			want:    true,
			wantErr: false,
		},
		{
			name:    "got error due to can NOT directly produce a value",
			expr:    `updated_at < '2024-04-05 23:59:59'`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queryexpr.GetQueryExprResult(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetQueryExprResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetQueryExprResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTreeNodeFromQueryExpr(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    any
		wantErr bool
	}{
		{
			name:    "success using func from the library",
			expr:    `findLast([1], # >= 1)`,
			want:    "findLast([1], # >= 1)",
			wantErr: false,
		},
		{
			name:    "success using equation",
			expr:    `false == !true`,
			want:    "false == !true",
			wantErr: false,
		},
		{
			name:    "got error using wrong syntax",
			expr:    `findLast(`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queryexpr.GetTreeNodeFromQueryExpr(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTreeNodeFromQueryExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.String() != tt.want {
					t.Errorf("GetTreeNodeFromQueryExpr() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
