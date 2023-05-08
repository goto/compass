package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/olivere/elastic/v7"
	"io"
	"strings"

	"github.com/goto/compass/core/asset"
)

const (
	defaultMaxResults                  = 200
	defaultGroupSize                   = 10
	defaultMinScore                    = 0.01
	defaultFunctionScoreQueryScoreMode = "sum"
	suggesterName                      = "name-phrase-suggest"
)

var returnedAssetFieldsResult = []string{"id", "urn", "type", "service", "name", "description", "data", "labels", "created_at", "updated_at"}

// Search the asset store
func (repo *DiscoveryRepository) Search(ctx context.Context, cfg asset.SearchConfig) (results []asset.SearchResult, err error) {
	if strings.TrimSpace(cfg.Text) == "" {
		err = asset.DiscoveryError{Err: errors.New("search text cannot be empty")}
		return
	}
	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	query, err := repo.buildQuery(cfg)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error building query %w", err)}
		return
	}

	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(defaultSearchIndex),
		repo.cli.client.Search.WithSize(maxResults),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithSourceIncludes(returnedAssetFieldsResult...),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error executing search %w", err)}
		return
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error decoding search response %w", err)}
		return
	}

	results = repo.toSearchResults(response.Hits.Hits)
	return
}

func (repo *DiscoveryRepository) Suggest(ctx context.Context, config asset.SearchConfig) (results []string, err error) {
	maxResults := config.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}

	query, err := repo.buildSuggestQuery(config)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error building query: %s", err)}
		return
	}
	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIndex(defaultSearchIndex),
		repo.cli.client.Search.WithSize(maxResults),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithContext(ctx),
	)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error executing search %w", err)}
		return
	}
	if res.IsError() {
		err = asset.DiscoveryError{Err: fmt.Errorf("error when searching %s", errorReasonFromResponse(res))}
		return
	}

	var response searchResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error decoding search response %w", err)}
		return
	}
	results, err = repo.toSuggestions(response)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error mapping response to suggestion %w", err)}
	}

	return
}

func (repo *DiscoveryRepository) buildQuery(cfg asset.SearchConfig) (io.Reader, error) {
	var query elastic.Query

	query = repo.buildTextQuery(cfg.Text)
	filterQueries := repo.buildFilterTermQueries(cfg.Filters)
	if filterQueries != nil {
		query = elastic.NewBoolQuery().Should(query).Filter(filterQueries...)
	}
	query = repo.buildFilterMatchQueries(query, cfg.Queries)
	query = repo.buildFunctionScoreQuery(query, cfg.RankBy)

	src, err := query.Source()
	if err != nil {
		return nil, err
	}

	payload := new(bytes.Buffer)
	q := &searchQuery{
		MinScore: defaultMinScore,
		Query:    src,
	}
	return payload, json.NewEncoder(payload).Encode(q)
}

func (repo *DiscoveryRepository) buildSuggestQuery(cfg asset.SearchConfig) (io.Reader, error) {
	suggester := elastic.NewCompletionSuggester(suggesterName).
		Field("name.suggest").
		SkipDuplicates(true).
		Size(5).
		Text(cfg.Text)
	src, err := elastic.NewSearchSource().
		Suggester(suggester).
		Source()
	if err != nil {
		return nil, fmt.Errorf("error building search source %w", err)
	}

	payload := new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(src)
	if err != nil {
		return payload, fmt.Errorf("error building reader %w", err)
	}

	return payload, err
}

func (repo *DiscoveryRepository) buildTextQuery(text string) elastic.Query {
	boostedFields := []string{
		"urn^10",
		"name^5",
	}

	return elastic.NewBoolQuery().
		Should(
			elastic.
				NewMultiMatchQuery(
					text,
					boostedFields...,
				),
			elastic.
				NewMultiMatchQuery(
					text,
					boostedFields...,
				).
				Fuzziness("AUTO"),
			elastic.
				NewMultiMatchQuery(
					text,
				).
				Fuzziness("AUTO"),
		)
}

func (repo *DiscoveryRepository) buildFilterMatchQueries(query elastic.Query, queries map[string]string) elastic.Query {
	if len(queries) == 0 {
		return query
	}

	esQueries := []elastic.Query{}
	for field, value := range queries {
		esQueries = append(esQueries,
			elastic.
				NewMatchQuery(field, value).
				Fuzziness("AUTO"))
	}

	return elastic.NewBoolQuery().
		Should(query).
		Filter(esQueries...)
}

func (repo *DiscoveryRepository) buildFilterTermQueries(filters map[string][]string) []elastic.Query {
	if len(filters) == 0 {
		return nil
	}

	var filterQueries []elastic.Query
	for key, rawValues := range filters {
		if len(rawValues) < 1 {
			continue
		}

		var values []interface{}
		for _, rawVal := range rawValues {
			values = append(values, rawVal)
		}

		key := fmt.Sprintf("%s.keyword", key)
		filterQueries = append(
			filterQueries,
			elastic.NewTermsQuery(key, values...),
		)
	}

	return filterQueries
}

func (repo *DiscoveryRepository) buildFilterExistsQueries(filters []string) []elastic.Query {
	if len(filters) == 0 {
		return nil
	}

	var filterQueries []elastic.Query
	for _, filterString := range filters {
		filterQueries = append(
			filterQueries,
			elastic.NewExistsQuery(fmt.Sprintf("%s.keyword", filterString)),
		)
	}

	return filterQueries
}

func (repo *DiscoveryRepository) buildFunctionScoreQuery(query elastic.Query, rankBy string) elastic.Query {
	if rankBy == "" {
		return query
	}

	factorFunc := elastic.NewFieldValueFactorFunction().
		Field(rankBy).
		Modifier("log1p").
		Missing(1.0).
		Weight(1.0)

	fsQuery := elastic.NewFunctionScoreQuery().
		ScoreMode(defaultFunctionScoreQueryScoreMode).
		AddScoreFunc(factorFunc).
		Query(query)

	return fsQuery
}

func (repo *DiscoveryRepository) toSearchResults(hits []searchHit) []asset.SearchResult {
	results := []asset.SearchResult{}
	for _, hit := range hits {
		r := hit.Source
		id := r.ID
		if id == "" { // this is for backward compatibility for asset without ID
			id = r.URN
		}
		results = append(results, asset.SearchResult{
			Type:        r.Type.String(),
			ID:          id,
			URN:         r.URN,
			Description: r.Description,
			Title:       r.Name,
			Service:     r.Service,
			Labels:      r.Labels,
			Data:        r.Data,
		})
	}
	return results
}

func (repo *DiscoveryRepository) toSuggestions(response searchResponse) (results []string, err error) {
	suggests, exists := response.Suggest[suggesterName]
	if !exists {
		err = errors.New("suggester key does not exist")
		return
	}
	results = []string{}
	for _, s := range suggests {
		for _, option := range s.Options {
			results = append(results, option.Text)
		}
	}

	return
}

func (repo *DiscoveryRepository) Group(ctx context.Context, cfg asset.GroupConfig) ([]asset.GroupResult, error) {
	if len(cfg.GroupBy) == 0 || cfg.GroupBy[0] == "" {
		err := asset.DiscoveryError{Err: fmt.Errorf("group by field cannot be empty")}
		return nil, err
	}
	query, err := repo.buildGroupQuery(cfg)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error building query %w", err)}
		return nil, err
	}
	cfg.Logger.Debug(fmt.Sprintf("group asset query %s for config %+v", query, cfg))

	res, err := repo.cli.client.Search(
		repo.cli.client.Search.WithFilterPath("aggregations"),
		repo.cli.client.Search.WithBody(query),
		repo.cli.client.Search.WithIgnoreUnavailable(true),
		repo.cli.client.Search.WithContext(ctx),
	)

	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error executing group query %w", err)}
		return nil, err
	}

	var response groupResponse

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error decoding group response %w", err)}
		return nil, err
	}

	groupResultMap := make(map[string][]asset.Asset)
	repo.ParseGroupResults(&response.Aggregations.TermAggregations, groupResultMap)
	results := repo.toGroupResults(groupResultMap)
	return results, nil
}

func (repo *DiscoveryRepository) ParseGroupResults(termAggregations *TermAggregations, groupResultMap map[string][]asset.Asset) {

	// Since we can have multiple groups, elastic search returns the results in nested order.
	// To make the structure flat we are clubbing the response based on one top group and appending the assets to it.
	// If it is not the top group, we recursively invoke the same function to fetch the nested assets.
	for _, bucket := range termAggregations.Buckets {
		if bucket.Group == nil {
			key := bucket.Key
			assetSource := bucket.Hits.Hits.Hits[0].Source
			if assets, ok := groupResultMap[key]; ok {
				groupResultMap[key] = append(assets, assetSource)
			} else {
				groupResultMap[key] = []asset.Asset{assetSource}
			}
		} else {
			repo.ParseGroupResults(bucket.Group, groupResultMap)
		}
	}
}

func (repo *DiscoveryRepository) toGroupResults(groupResultMap map[string][]asset.Asset) []asset.GroupResult {
	groupResult := make([]asset.GroupResult, len(groupResultMap))
	idx := 0
	for key, assets := range groupResultMap {
		groupResult[idx].Key = key
		groupResult[idx].Assets = make([]asset.Asset, len(assets))
		groupResult[idx].Assets = assets
		idx++
	}
	return groupResult
}

type GroupQuery struct {
	Query        interface{} `json:"query,omitempty"`
	Aggregations interface{} `json:"aggregations"`
}

func (repo *DiscoveryRepository) buildGroupQuery(cfg asset.GroupConfig) (io.Reader, error) {

	var querySource interface{}

	size := cfg.Size
	if size <= 0 {
		size = defaultGroupSize
	}

	// This code takes care of creating filter term queries from the input filters mentioned in request.
	filterQueries := repo.buildFilterExistsQueries(cfg.GroupBy)
	filterQueries = append(filterQueries, repo.buildFilterTermQueries(cfg.Filters)...)
	if filterQueries != nil {
		querySource, _ = elastic.NewBoolQuery().Filter(filterQueries...).Source()
	}

	// By default, the groupby fields would be part of the response hence added them in the input included fields list.
	includedFields := cfg.GroupBy
	if len(cfg.IncludedFields) > 0 {
		includedFields = append(cfg.GroupBy, cfg.IncludedFields...)
	}

	// Hits aggregation helps to return the specific parts of _source in response.
	fetchSourceContext := elastic.NewFetchSourceContext(true).Include(includedFields...)
	searchSource := elastic.NewSearchSource().FetchSourceContext(fetchSourceContext)
	hitsAggregations := elastic.NewTopHitsAggregation().SearchSource(searchSource)

	// Terms Aggregation helps us to group by based on particular fields.
	// If multiple fields are provided in group by a sub aggregation of type Term Aggregation is added.
	var termAggregations *elastic.TermsAggregation
	for idx, group := range cfg.GroupBy {
		if idx == 0 {
			termAggregations = elastic.NewTermsAggregation().Field(fmt.Sprintf("%s.keyword", group)).Size(size).
				SubAggregation("hits", hitsAggregations)
		} else {
			termAggregations = elastic.NewTermsAggregation().Field(fmt.Sprintf("%s.keyword", group)).Size(size).
				SubAggregation("group", termAggregations)
		}
	}

	// A term aggregation is also added for each field provided in included fields because every field in response also
	// needs to be part of aggregation.
	for _, field := range cfg.IncludedFields {
		termAggregations = elastic.NewTermsAggregation().Field(fmt.Sprintf("%s.keyword", field)).Size(size).
			SubAggregation("group", termAggregations)
	}

	aggregations := elastic.Aggregations{}
	src, _ := termAggregations.Source()
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(src)
	if err != nil {
		err = asset.DiscoveryError{Err: fmt.Errorf("error encoding group request %w", err)}
		return nil, err
	}
	aggregations["term_agg"] = payload.Bytes()

	payload = new(bytes.Buffer)
	q := &GroupQuery{
		Query:        querySource,
		Aggregations: aggregations,
	}

	return payload, json.NewEncoder(payload).Encode(q)
}
