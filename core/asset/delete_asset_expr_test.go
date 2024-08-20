package asset_test

import (
	"errors"
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/stretchr/testify/assert"
)

func TestDeleteAssetExpr_ToQuery(t *testing.T) {
	queryExp := `name == "John" || service not in ["test1","test2","test3"]`
	sqlExpr := queryexpr.SQLExpr(queryExp)
	esExpr := queryexpr.ESExpr(queryExp)
	wrongExpr := queryexpr.SQLExpr("findLast(")
	tests := []struct {
		name    string
		exprStr queryexpr.ExprStr
		want    string
		wantErr bool
	}{
		{
			name: "convert to SQL query",
			exprStr: asset.DeleteAssetExpr{
				ExprStr: &sqlExpr,
			},
			want:    "((name = 'John') OR (service NOT IN ('test1', 'test2', 'test3')))",
			wantErr: false,
		},
		{
			name: "convert to ES query",
			exprStr: asset.DeleteAssetExpr{
				ExprStr: &esExpr,
			},
			want:    `{"query":{"bool":{"should":[{"term":{"name":"John"}},{"bool":{"must_not":[{"terms":{"service.keyword":["test1","test2","test3"]}}]}}]}}}`,
			wantErr: false,
		},
		{
			name: "got error due to wrong syntax",
			exprStr: asset.DeleteAssetExpr{
				ExprStr: &wrongExpr,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := asset.DeleteAssetExpr{
				ExprStr: tt.exprStr,
			}
			got, err := d.ToQuery()
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

func TestDeleteAssetExpr_Validate(t *testing.T) {
	tests := []struct {
		name      string
		exprStrFn func() queryexpr.ExprStr
		expectErr error
		wantErr   bool
	}{
		{
			name: "error get identifiers map",
			exprStrFn: func() queryexpr.ExprStr {
				wrongExpr := queryexpr.SQLExpr("findLast(")
				return asset.DeleteAssetExpr{
					ExprStr: &wrongExpr,
				}
			},
			expectErr: errors.New("error parsing expression"),
			wantErr:   true,
		},
		{
			name: "error miss refreshed_at not exist",
			exprStrFn: func() queryexpr.ExprStr {
				missRefreshedAt := queryexpr.SQLExpr(`updated_at < "2023-12-12 23:59:59" && type == "table" && service in ["test1","test2","test3"]`)
				return asset.DeleteAssetExpr{
					ExprStr: &missRefreshedAt,
				}
			},
			expectErr: errors.New("must exists these identifiers: refreshed_at, type, and service"),
			wantErr:   true,
		},
		{
			name: "error miss type not exist",
			exprStrFn: func() queryexpr.ExprStr {
				missType := queryexpr.SQLExpr(`refreshed_at < "2023-12-12 23:59:59" && service in ["test1","test2","test3"]`)
				return asset.DeleteAssetExpr{
					ExprStr: &missType,
				}
			},
			expectErr: errors.New("must exists these identifiers: refreshed_at, type, and service"),
			wantErr:   true,
		},
		{
			name: "error miss service not exist",
			exprStrFn: func() queryexpr.ExprStr {
				missService := queryexpr.SQLExpr(`refreshed_at < "2023-12-12 23:59:59" && type == "table"`)
				return asset.DeleteAssetExpr{
					ExprStr: &missService,
				}
			},
			expectErr: errors.New("must exists these identifiers: refreshed_at, type, and service"),
			wantErr:   true,
		},
		{
			name: "error wrong operator for type identifier",
			exprStrFn: func() queryexpr.ExprStr {
				wrongTypeOperator := queryexpr.SQLExpr(`refreshed_at < "2023-12-12 23:59:59" && type != "table" && service in ["test1","test2","test3"]`)
				return asset.DeleteAssetExpr{
					ExprStr: &wrongTypeOperator,
				}
			},
			expectErr: errors.New("identifier type and service must be equals (==) or IN operator"),
			wantErr:   true,
		},
		{
			name: "error wrong operator for service identifier",
			exprStrFn: func() queryexpr.ExprStr {
				wrongServiceOperator := queryexpr.SQLExpr(`refreshed_at < "2023-12-12 23:59:59" && type != "table" && service not in ["test1","test2","test3"]`)
				return asset.DeleteAssetExpr{
					ExprStr: &wrongServiceOperator,
				}
			},
			expectErr: errors.New("identifier type and service must be equals (==) or IN operator"),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.exprStrFn().Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				assert.ErrorContains(t, err, tt.expectErr.Error())
			}
		})
	}
}
