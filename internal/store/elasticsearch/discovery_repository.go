package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/goto/salt/log"
)

// DiscoveryRepository implements discovery.Repository
// with elasticsearch as the backing store.
type DiscoveryRepository struct {
	cli                       *Client
	logger                    log.Logger
	requestTimeout            time.Duration
	columnSearchExclusionList []string
}

func NewDiscoveryRepository(cli *Client, logger log.Logger, requestTimeout time.Duration, colSearchExclusionList []string) *DiscoveryRepository {
	return &DiscoveryRepository{
		cli:                       cli,
		logger:                    logger,
		requestTimeout:            requestTimeout,
		columnSearchExclusionList: colSearchExclusionList,
	}
}

func (repo *DiscoveryRepository) createIndexIfNotExists(ctx context.Context, discoveryOp, indexName, alias string) error {
	idxExists, err := repo.cli.indexExists(ctx, discoveryOp, indexName)
	if err != nil {
		return asset.DiscoveryError{
			Op:    "IndexExists",
			Index: indexName,
			Err:   err,
		}
	}

	if !idxExists {
		if err := repo.cli.CreateIdx(ctx, discoveryOp, indexName, alias); err != nil {
			var de asset.DiscoveryError
			if ok := errors.As(err, &de); ok {
				if de.ESCode == "resource_already_exists_exception" {
					repo.logger.Warn(de.Err.Error())
					return nil
				}
			}

			return asset.DiscoveryError{
				Op:    "CreateIndex",
				Index: indexName,
				Err:   err,
			}
		}
	}

	return nil
}

func (repo *DiscoveryRepository) Upsert(ctx context.Context, ast asset.Asset) error {
	if ast.ID == "" {
		return asset.ErrEmptyID
	}
	if !ast.Type.IsValid() {
		return fmt.Errorf("type [%s] is invalid: %w", ast.Type, asset.ErrUnknownType)
	}

	if err := repo.createIndexIfNotExists(ctx, "Upsert", ast.Service, defaultSearchIndex); err != nil {
		return err
	}

	return repo.indexAsset(ctx, ast)
}

func (repo *DiscoveryRepository) SyncAssets(ctx context.Context, indexName string) (func() error, error) {
	backupIndexName := fmt.Sprintf("%+v-bak", indexName)

	err := repo.updateIndexSettings(ctx, indexName, `{"settings":{"index.blocks.write":true}}`)
	if err != nil {
		return nil, err
	}

	err = repo.clone(ctx, indexName, backupIndexName)
	if err != nil {
		return nil, err
	}

	err = repo.updateAlias(ctx, backupIndexName, defaultSearchIndex)
	if err != nil {
		return nil, err
	}

	err = repo.deleteByIndexName(ctx, indexName)
	if err != nil {
		return nil, err
	}

	err = repo.createIndexIfNotExists(ctx, "SyncAssets", indexName, "")
	if err != nil {
		return nil, err
	}

	cleanupFn := func() error {
		err = repo.updateAlias(ctx, indexName, defaultSearchIndex)
		if err != nil {
			return err
		}

		err = repo.deleteByIndexName(ctx, backupIndexName)
		if err != nil {
			return err
		}

		err = repo.updateIndexSettings(ctx, indexName, `{"settings":{"index.blocks.write":false}}`)
		if err != nil {
			return err
		}
		return nil
	}

	return cleanupFn, err
}

func (repo *DiscoveryRepository) DeleteByID(ctx context.Context, assetID string) error {
	if assetID == "" {
		return asset.ErrEmptyID
	}

	return repo.deleteWithQuery(ctx, "DeleteByID", fmt.Sprintf(`{"query":{"term":{"_id": %q}}}`, assetID))
}

func (repo *DiscoveryRepository) DeleteByURN(ctx context.Context, assetURN string) error {
	if assetURN == "" {
		return asset.ErrEmptyURN
	}

	return repo.deleteWithQuery(ctx, "DeleteByURN", fmt.Sprintf(`{"query":{"term":{"urn.keyword": %q}}}`, assetURN))
}

func (repo *DiscoveryRepository) SoftDeleteByURN(ctx context.Context, softDeleteAsset asset.SoftDeleteAsset) error {
	if softDeleteAsset.URN == "" {
		return asset.ErrEmptyURN
	}

	return repo.softDeleteAsset(ctx, "DeleteByURN", softDeleteAsset)
}

func (repo *DiscoveryRepository) DeleteByQueryExpr(ctx context.Context, queryExpr queryexpr.ExprStr) error {
	if strings.TrimSpace(queryExpr.String()) == "" {
		return asset.ErrEmptyQuery
	}

	esQuery, err := queryexpr.ValidateAndGetQueryFromExpr(queryExpr)
	if err != nil {
		return err
	}

	return repo.deleteWithQuery(ctx, "DeleteByQueryExpr", esQuery)
}

func (repo *DiscoveryRepository) SoftDeleteByQueryExpr(ctx context.Context, softDeleteAssets asset.SoftDeleteAssets) error {
	if strings.TrimSpace(softDeleteAssets.QueryExpr.String()) == "" {
		return asset.ErrEmptyQuery
	}

	esQuery, err := queryexpr.ValidateAndGetQueryFromExpr(softDeleteAssets.QueryExpr)
	if err != nil {
		return err
	}
	queryMap, err := queryexpr.QueryStringToMap(esQuery)

	// Create the update request body
	bodyRequest := map[string]interface{}{
		"query": queryMap["query"],
		"script": map[string]interface{}{
			"source": `
                ctx._source.is_deleted = true;
                ctx._source.updated_at = params.updated_at;
                ctx._source.refreshed_at = params.refreshed_at;
                ctx._source.updated_by = params.updated_by
            `,
			"lang": "painless",
			"params": map[string]interface{}{
				"updated_at":   softDeleteAssets.UpdatedAt,
				"refreshed_at": softDeleteAssets.RefreshedAt,
				"updated_by":   softDeleteAssets.UpdatedBy,
			},
		},
	}

	return repo.softDeleteAsset(ctx, "DeleteByQueryExpr", bodyRequest)
}

func (repo *DiscoveryRepository) deleteWithQuery(ctx context.Context, discoveryOp, qry string) (err error) {
	defer func(start time.Time) {
		const op = "delete_by_query"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: discoveryOp,
			start:       start,
			err:         err,
		})
	}(time.Now())

	deleteByQ := repo.cli.client.DeleteByQuery
	res, err := deleteByQ(
		[]string{defaultSearchIndex},
		strings.NewReader(qry),
		deleteByQ.WithContext(ctx),
		deleteByQ.WithRefresh(true),
		deleteByQ.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return asset.DiscoveryError{
			Op:  "DeleteDoc",
			Err: fmt.Errorf("query: %s: %w", qry, err),
		}
	}

	defer drainBody(res)
	if res.IsError() {
		code, reason := errorCodeAndReason(res)
		return asset.DiscoveryError{
			Op:     "DeleteDoc",
			ESCode: code,
			Err:    fmt.Errorf("query: %s: %s", qry, reason),
		}
	}

	return nil
}

func (repo *DiscoveryRepository) softDeleteAsset(ctx context.Context, discoveryOp string, softDeleteAsset asset.SoftDeleteAsset) (err error) {
	defer func(start time.Time) {
		const op = "soft_delete_by_query"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: discoveryOp,
			start:       start,
			err:         err,
		})
	}(time.Now())

	// First get the current version
	currentVersion, err := repo.GetCurrentAssetVersion(ctx, softDeleteAsset.URN, 2*time.Second)
	if err != nil {
		return asset.DiscoveryError{
			Op:  "GetCurrentVersion",
			Err: fmt.Errorf("failed to get current version for URN %s: %w", softDeleteAsset.URN, err),
		}
	}
	newVersion, err := asset.IncreaseMinorVersion(currentVersion)
	if err != nil {
		return err
	}

	// Create the update request body
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"urn.keyword": softDeleteAsset.URN,
			},
		},
		"script": map[string]interface{}{
			"source": `
                ctx._source.is_deleted = true;
                ctx._source.updated_at = params.updated_at;
                ctx._source.refreshed_at = params.refreshed_at;
                ctx._source.updated_by = params.updated_by;
                ctx._source.version = params.version;
            `,
			"lang": "painless",
			"params": map[string]interface{}{
				"updated_at":   softDeleteAsset.UpdatedAt,
				"refreshed_at": softDeleteAsset.RefreshedAt,
				"updated_by":   user.User{ID: softDeleteAsset.UpdatedBy},
				"version":      newVersion,
			},
		},
	}

	buf, err := encodeBodyRequest(body)
	if err != nil {
		return asset.DiscoveryError{
			Op:  "SoftDeleteByURN",
			Err: err,
		}
	}

	// Execute UpdateByQuery
	res, err := repo.cli.client.UpdateByQuery(
		[]string{defaultSearchIndex},
		repo.cli.client.UpdateByQuery.WithContext(ctx),
		repo.cli.client.UpdateByQuery.WithBody(buf),
		repo.cli.client.UpdateByQuery.WithRefresh(true),
		repo.cli.client.UpdateByQuery.WithIgnoreUnavailable(true),
		repo.cli.client.UpdateByQuery.WithWaitForCompletion(true),
		repo.cli.client.UpdateByQuery.WithConflicts("proceed"),
	)
	if err != nil {
		return asset.DiscoveryError{
			Op:  "DeleteDoc",
			Err: fmt.Errorf("urn: %s: %w", softDeleteAsset.URN, err),
		}
	}

	defer drainBody(res)
	if res.IsError() {
		code, reason := errorCodeAndReason(res)
		return asset.DiscoveryError{
			Op:     "DeleteDoc",
			ESCode: code,
			Err:    fmt.Errorf("urn: %s: %s", softDeleteAsset.URN, reason),
		}
	}

	return nil
}

// GetCurrentAssetVersion is helper function to get current version. Used in soft delete func and tests.
func (repo *DiscoveryRepository) GetCurrentAssetVersion(
	ctx context.Context,
	urn string,
	timeout time.Duration,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"urn.keyword": urn,
			},
		},
		"size":    1,
		"_source": []string{"version"},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return "", fmt.Errorf("failed to encode query: %w", err)
	}

	// Retry loop every 500ms
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Timeout reached
			return "", fmt.Errorf("asset %s not found after %v", urn, timeout)

		case <-ticker.C:
			// Execute search
			res, err := repo.cli.client.Search(
				repo.cli.client.Search.WithContext(ctx),
				repo.cli.client.Search.WithIndex(defaultSearchIndex),
				repo.cli.client.Search.WithBody(&buf),
				repo.cli.client.Search.WithIgnoreUnavailable(true),
			)
			if err != nil {
				continue // Retry on network errors
			}

			// Process response if no error
			var result struct {
				Hits struct {
					Hits []struct {
						Source struct {
							Version string `json:"version"`
						} `json:"_source"`
					} `json:"hits"`
				} `json:"hits"`
			}

			if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
				res.Body.Close()
				continue // Retry on decode errors
			}
			res.Body.Close()

			if len(result.Hits.Hits) > 0 {
				return result.Hits.Hits[0].Source.Version, nil // Found!
			}
		}
	}
}

func (repo *DiscoveryRepository) indexAsset(ctx context.Context, ast asset.Asset) (err error) {
	defer func(start time.Time) {
		const op = "index"
		repo.cli.instrumentOp(ctx, instrumentParams{
			op:          op,
			discoveryOp: "Upsert",
			start:       start,
			err:         err,
		})
	}(time.Now())

	body, err := encodeBodyRequest(ast)
	if err != nil {
		return asset.DiscoveryError{
			Op:  "EncodeAsset",
			ID:  ast.ID,
			Err: err,
		}
	}

	index := repo.cli.client.Index
	resp, err := index(
		ast.Service,
		body,
		index.WithDocumentID(url.PathEscape(ast.ID)),
		index.WithContext(ctx),
	)
	if err != nil {
		return asset.DiscoveryError{
			Op:    "IndexDoc",
			ID:    ast.ID,
			Index: ast.Service,
			Err:   err,
		}
	}
	defer drainBody(resp)

	if resp.IsError() {
		code, reason := errorCodeAndReason(resp)
		return asset.DiscoveryError{
			Op:     "IndexDoc",
			ID:     ast.ID,
			Index:  ast.Service,
			ESCode: code,
			Err:    errors.New(reason),
		}
	}

	return nil
}

func encodeBodyRequest(body interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}

	return &buf, nil
}

func (repo *DiscoveryRepository) clone(ctx context.Context, indexName, clonedIndexName string) error {
	idxExists, err := repo.cli.indexExists(ctx, "CloneIndex", clonedIndexName)
	if err != nil {
		return asset.DiscoveryError{
			Op:    "IndexExists",
			Index: indexName,
			Err:   err,
		}
	}
	if idxExists {
		return nil // skip clone when backup already created
	}

	cloneFn := repo.cli.client.Indices.Clone
	resp, err := cloneFn(indexName, clonedIndexName, cloneFn.WithContext(ctx))
	if err != nil {
		return asset.DiscoveryError{
			Op:    "CloneIndex",
			Index: indexName,
			Err:   err,
		}
	}

	if resp.IsError() {
		code, reason := errorCodeAndReason(resp)
		return asset.DiscoveryError{
			Op:     "CloneIndex",
			Index:  indexName,
			ESCode: code,
			Err:    errors.New(reason),
		}
	}

	return nil
}

func (repo *DiscoveryRepository) updateAlias(ctx context.Context, indexName, alias string) error {
	putAliasFn := repo.cli.client.Indices.PutAlias
	resp, err := putAliasFn([]string{indexName}, alias, putAliasFn.WithContext(ctx))
	if err != nil {
		return asset.DiscoveryError{
			Op:    "UpdateAlias",
			Index: indexName,
			Err:   err,
		}
	}

	if resp.IsError() {
		code, reason := errorCodeAndReason(resp)
		return asset.DiscoveryError{
			Op:     "UpdateAlias",
			Index:  indexName,
			ESCode: code,
			Err:    errors.New(reason),
		}
	}
	return nil
}

func (repo *DiscoveryRepository) deleteByIndexName(ctx context.Context, indexName string) error {
	deleteFn := repo.cli.client.Indices.Delete
	resp, err := deleteFn([]string{indexName}, deleteFn.WithContext(ctx))
	if err != nil {
		return asset.DiscoveryError{
			Op:    "DeleteIndex",
			Index: indexName,
			Err:   err,
		}
	}

	if resp.IsError() {
		code, reason := errorCodeAndReason(resp)
		return asset.DiscoveryError{
			Op:     "DeleteIndex",
			Index:  indexName,
			ESCode: code,
			Err:    errors.New(reason),
		}
	}

	return nil
}

func (repo *DiscoveryRepository) updateIndexSettings(ctx context.Context, indexName, body string) error {
	putSettings := repo.cli.client.Indices.PutSettings

	resp, err := putSettings(strings.NewReader(body),
		putSettings.WithIndex(indexName),
		putSettings.WithContext(ctx))
	if err != nil {
		return asset.DiscoveryError{
			Op:    "UpdateSettings",
			Index: indexName,
			Err:   err,
		}
	}

	if resp.IsError() {
		code, reason := errorCodeAndReason(resp)
		return asset.DiscoveryError{
			Op:     "UpdateSettings",
			Index:  indexName,
			ESCode: code,
			Err:    errors.New(reason),
		}
	}

	return err
}
