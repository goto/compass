package queryexpr_test

import (
	"reflect"
	"testing"

	"github.com/goto/compass/pkg/queryexpr"
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
