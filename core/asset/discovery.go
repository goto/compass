package asset

//go:generate mockery --name=DiscoveryRepository -r --case underscore --with-expecter --structname DiscoveryRepository --filename discovery_repository.go --output=./mocks
import (
	"context"

	"github.com/goto/compass/pkg/queryexpr"
)

type DiscoveryRepository interface {
	Upsert(context.Context, Asset) error
	DeleteByID(ctx context.Context, assetID string) error
	DeleteByURN(ctx context.Context, assetURN string) error
	SoftDeleteByURN(ctx context.Context, softDeleteAsset SoftDeleteAsset) error
	DeleteByQueryExpr(ctx context.Context, queryExpr queryexpr.ExprStr) error
	SoftDeleteByQueryExpr(ctx context.Context, softDeleteAssets SoftDeleteAssets) error
	Search(ctx context.Context, cfg SearchConfig) (results []SearchResult, err error)
	Suggest(ctx context.Context, cfg SearchConfig) (suggestions []string, err error)
	GroupAssets(ctx context.Context, cfg GroupConfig) (results []GroupResult, err error)
	SyncAssets(ctx context.Context, indexName string) (cleanup func() error, err error)
}

// GroupConfig represents a group query along
// with any corresponding filter(s)
type GroupConfig struct {
	// IncludedFields specifies the fields to return in response
	IncludedFields []string

	// GroupBy fields to group on
	GroupBy []string

	// Filters specifies document level values to look for.
	// Multiple values can be specified for a single key
	Filters SearchFilter

	// Number of documents you want in response
	Size int
}

// SearchFilter is a filter intended to be used as a search
// criteria for operations involving asset search
type SearchFilter = map[string][]string

// SearchFlags is intended to be used as flags to control the search
// behavior (e.g. column level search, disable fuzzy, enable highlight etc) for
// operations involving asset search
type SearchFlags struct {
	EnableHighlight bool

	// DisableFuzzy disables fuzziness on search
	DisableFuzzy bool

	IsColumnSearch bool
}

// SearchConfig represents a search query along
// with any corresponding filter(s)
type SearchConfig struct {
	// Text to search for
	Text string

	// Filters specifies document level values to look for.
	// Multiple values can be specified for a single key
	Filters SearchFilter

	// Number of relevant results to return
	MaxResults int

	// RankBy is a param to rank based on a specific parameter
	RankBy string

	// Queries is a param to search a resource based on asset's fields
	Queries map[string]string

	// Flags flags to control the search behavior (e.g. column level search, disable fuzzy, etc)
	Flags SearchFlags

	// Offset parameter defines the offset from the first result you want to fetch
	// Note that MaxResults + Offset can not be more than the `index.max_result_window` index setting in ES cluster, which defaults to 10,000
	Offset int

	// IncludeFields specifies the fields to return in response
	IncludeFields []string
}

// SearchResult represents an item/result in a list of search results
type SearchResult struct {
	ID          string                 `json:"id"`
	URN         string                 `json:"urn"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Service     string                 `json:"service"`
	Description string                 `json:"description"`
	Labels      map[string]string      `json:"labels"`
	Data        map[string]interface{} `json:"data"`
}

type GroupResult struct {
	Fields []GroupField
	Assets []Asset
}

type GroupField struct {
	Name  string
	Value string
}

// ToAsset returns search result as asset
func (sr SearchResult) ToAsset() Asset {
	return Asset{
		ID:          sr.ID,
		URN:         sr.URN,
		Name:        sr.Title,
		Type:        Type(sr.Type),
		Service:     sr.Service,
		Description: sr.Description,
		Labels:      sr.Labels,
		Data:        sr.Data,
	}
}
