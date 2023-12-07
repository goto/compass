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

func (repo *DiscoveryRepository) createIndex(ctx context.Context, discoveryOp, indexName, alias string) error {
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
		return asset.ErrUnknownType
	}

	if err := repo.createIndex(ctx, "Upsert", ast.Service, defaultSearchIndex); err != nil {
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

	err = repo.createIndex(ctx, "SyncAssets", indexName, "")
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
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

	return cleanup, err
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

	body, err := createUpsertBody(ast)
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

func createUpsertBody(ast asset.Asset) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(ast); err != nil {
		return nil, fmt.Errorf("encode asset: %w", err)
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
