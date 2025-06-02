package asset

//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname AssetRepository --filename asset_repository.go --output=./mocks
import (
	"context"
	"time"

	"github.com/goto/compass/core/user"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/r3labs/diff/v2"
)

type Repository interface {
	GetAll(context.Context, Filter) ([]Asset, error)
	GetCount(context.Context, Filter) (int, error)
	GetCountByQueryExpr(ctx context.Context, queryExpr queryexpr.ExprStr) (int, error)
	GetByID(ctx context.Context, id string) (Asset, error)
	GetByURN(ctx context.Context, urn string) (Asset, error)
	GetVersionHistory(ctx context.Context, flt Filter, id string) ([]Asset, error)
	GetByVersionWithID(ctx context.Context, id, version string) (Asset, error)
	GetByVersionWithURN(ctx context.Context, urn, version string) (Asset, error)
	GetTypes(ctx context.Context, flt Filter) (map[Type]int, error)
	Upsert(ctx context.Context, ast *Asset) (string, error)
	UpsertPatch(ctx context.Context, ast *Asset, patchData map[string]interface{}) (string, error)
	DeleteByID(ctx context.Context, id string) (string, error)
	DeleteByURN(ctx context.Context, urn string) error
	SoftDeleteByID(ctx context.Context, id string, softDeleteAsset SoftDeleteAsset) (string, error)
	SoftDeleteByURN(ctx context.Context, urn string, softDeleteAsset SoftDeleteAsset) error
	DeleteByQueryExpr(ctx context.Context, queryExpr queryexpr.ExprStr) ([]string, error)
	SoftDeleteByQueryExpr(ctx context.Context, softDeleteAssetsByQueryExpr SoftDeleteAssetsByQueryExpr) error
	AddProbe(ctx context.Context, assetURN string, probe *Probe) error
	GetProbes(ctx context.Context, assetURN string) ([]Probe, error)
	GetProbesWithFilter(ctx context.Context, flt ProbesFilter) (map[string][]Probe, error)
}

// Asset is a model that wraps arbitrary data with Compass' context
type Asset struct {
	ID          string                 `json:"id" diff:"-"`
	URN         string                 `json:"urn" diff:"-"`
	Type        Type                   `json:"type" diff:"-"`
	Service     string                 `json:"service" diff:"-"`
	Name        string                 `json:"name" diff:"name"`
	Description string                 `json:"description" diff:"description"`
	Data        map[string]interface{} `json:"data" diff:"data"`
	URL         string                 `json:"url" diff:"url"`
	Labels      map[string]string      `json:"labels" diff:"labels"`
	Owners      []user.User            `json:"owners,omitempty" diff:"owners"`
	CreatedAt   time.Time              `json:"created_at" diff:"-"`
	UpdatedAt   time.Time              `json:"updated_at" diff:"-"`
	RefreshedAt *time.Time             `json:"refreshed_at" diff:"-"`
	Version     string                 `json:"version" diff:"-"`
	UpdatedBy   user.User              `json:"updated_by" diff:"-"`
	IsDeleted   bool                   `json:"is_deleted" diff:"is_deleted"`
	Changelog   diff.Changelog         `json:"changelog,omitempty" diff:"-"`
	Probes      []Probe                `json:"probes,omitempty"`
}

type SoftDeleteAsset struct {
	URN         string         `json:"urn"`
	UpdatedAt   time.Time      `json:"updated_at"`
	RefreshedAt time.Time      `json:"refreshed_at"`
	Version     string         `json:"version"`
	UpdatedBy   string         `json:"updated_by"`
	IsDeleted   bool           `json:"is_deleted"`
	Changelog   diff.Changelog `json:"changelog,omitempty"`
}

func NewSoftDeleteAsset(
	updatedAt, refreshedAt time.Time,
	updatedBy string,
) SoftDeleteAsset {
	return SoftDeleteAsset{
		UpdatedAt:   updatedAt,
		RefreshedAt: refreshedAt,
		UpdatedBy:   updatedBy,
		IsDeleted:   true,
		Changelog: diff.Changelog{
			{
				Type: "delete",
				Path: []string{"is_deleted"},
				From: false,
				To:   true,
			},
		},
	}
}

type SoftDeleteAssetsByQueryExpr struct {
	UpdatedAt    time.Time         `json:"updated_at"`
	RefreshedAt  time.Time         `json:"refreshed_at"`
	UpdatedBy    string            `json:"updated_by"`
	IsDeleted    bool              `json:"is_deleted"`
	Changelog    diff.Changelog    `json:"changelog,omitempty"`
	QueryExprStr string            `json:"query_expr"`
	QueryExpr    queryexpr.ExprStr `json:"-"`
}

func NewSoftDeleteAssetsByQueryExpr(
	updatedAt, refreshedAt time.Time,
	updatedBy, queryExprStr string,
	queryExpr queryexpr.ExprStr,
) SoftDeleteAssetsByQueryExpr {
	return SoftDeleteAssetsByQueryExpr{
		UpdatedAt:    updatedAt,
		RefreshedAt:  refreshedAt,
		UpdatedBy:    updatedBy,
		QueryExprStr: queryExprStr,
		QueryExpr:    queryExpr,
		IsDeleted:    true,
		Changelog: diff.Changelog{
			{
				Type: "delete",
				Path: []string{"is_deleted"},
				From: false,
				To:   true,
			},
		},
	}
}

// Diff returns nil changelog with nil error if equal
// returns wrapped r3labs/diff Changelog struct with nil error if not equal
func (a *Asset) Diff(otherAsset *Asset) (diff.Changelog, error) {
	return diff.Diff(a, otherAsset, diff.DiscardComplexOrigin(), diff.AllowTypeMismatch(true))
}

// Patch appends asset with data from map. It mutates the asset itself.
// It is using json annotation of the struct to patch the correct keys
func (a *Asset) Patch(patchData map[string]interface{}) {
	patchAsset(a, patchData)
}
