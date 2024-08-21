package queryexpr_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/goto/compass/pkg/queryexpr"
)

func TestSQLExpr_String(t *testing.T) {
	tests := []struct {
		expr queryexpr.SQLExpr
		want string
	}{
		{
			expr: queryexpr.SQLExpr("test"),
			want: "test",
		},
		{
			expr: queryexpr.SQLExpr("bool_identifier == !(findLast([1, 2, 3, 4], # > 2) == 4)"),
			want: "bool_identifier == !(findLast([1, 2, 3, 4], # > 2) == 4)",
		},
	}
	for i, tt := range tests {
		t.Run("test-case-"+string(rune(i)), func(t *testing.T) {
			if got := tt.expr.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLExpr_ToQuery(t *testing.T) {
	tests := []struct {
		name    string
		expr    queryexpr.SQLExpr
		want    string
		wantErr bool
	}{
		{
			name:    "less than condition with single quote",
			expr:    queryexpr.SQLExpr(`updated_at < '2024-04-05 23:59:59'`),
			want:    `(updated_at < '2024-04-05 23:59:59')`,
			wantErr: false,
		},
		{
			name:    "greater than condition with double quote",
			expr:    queryexpr.SQLExpr(`updated_at > "2024-04-05 23:59:59"`),
			want:    `(updated_at > '2024-04-05 23:59:59')`,
			wantErr: false,
		},
		{
			name:    "in condition",
			expr:    queryexpr.SQLExpr(`service in ["test1","test2","test3"]`),
			want:    `(service IN ('test1', 'test2', 'test3'))`,
			wantErr: false,
		},
		{
			name:    "equals or not in condition",
			expr:    queryexpr.SQLExpr(`name == "John" || service not in ["test1","test2","test3"]`),
			want:    `((name = 'John') OR (service NOT IN ('test1', 'test2', 'test3')))`,
			wantErr: false,
		},
		{
			name:    "complex query expression that can directly produce a value",
			expr:    queryexpr.SQLExpr(`(bool_identifier == !(findLast([1, 2, 3, 4], # > 2) == 4)) && name != 'John'`),
			want:    `((bool_identifier = false) AND (name != 'John'))`,
			wantErr: false,
		},
		{
			name:    "complex query expression that can directly produce a value regarding time",
			expr:    queryexpr.SQLExpr(`refreshed_at <= (now() - duration('1h'))`),
			want:    fmt.Sprintf("(refreshed_at <= '%s')", time.Now().Add(-1*time.Hour).Format(time.RFC3339)),
			wantErr: false,
		},
		{
			name:    "complex query expression that can NOT directly produce a value",
			expr:    queryexpr.SQLExpr(`service in filter(assets, .Service startsWith "T")`),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.expr.ToQuery()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToQuery() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLExpr_Validate(t *testing.T) {
	t.Run("should return nil as default validation", func(t *testing.T) {
		expr := queryexpr.SQLExpr("query_sql == 'test'")
		if err := (&expr).Validate(); err != nil {
			t.Errorf("Validate() error = %v, wantErr %v", err, nil)
		}
	})
}
