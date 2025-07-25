package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/r3labs/diff/v2"
)

var (
	errOffsetCannotBeNegative = errors.New("offset cannot be negative")
	errSizeCannotBeNegative   = errors.New("size cannot be negative")
)

// AssetRepository is a type that manages user operation to the primary database
type AssetRepository struct {
	client              *Client
	userRepo            *UserRepository
	defaultGetMaxSize   int
	defaultUserProvider string
}

// GetAll retrieves list of assets with filters
func (r *AssetRepository) GetAll(ctx context.Context, flt asset.Filter) ([]asset.Asset, error) {
	if flt.Offset < 0 {
		return nil, errOffsetCannotBeNegative
	}
	if flt.Size < 0 {
		return nil, errSizeCannotBeNegative
	}

	builder := r.getAssetSQLWithIsDeleted(flt.IsDeleted, true).Offset(uint64(flt.Offset))
	size := flt.Size

	if size > 0 {
		builder = r.getAssetSQLWithIsDeleted(flt.IsDeleted, true).Limit(uint64(size)).Offset(uint64(flt.Offset))
	}
	builder = r.BuildFilterQuery(builder, flt)
	builder = r.buildOrderQuery(builder, flt)
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var ams []*AssetModel
	err = r.client.db.SelectContext(ctx, &ams, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting asset list: %w", err)
	}

	assets := make([]asset.Asset, len(ams))
	for i, am := range ams {
		assets[i] = am.toAsset(nil)
	}

	return assets, nil
}

// GetTypes fetches types with assets count for all available types
// and returns them as a map[typeName]count
func (r *AssetRepository) GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error) {
	builder := r.getAssetsGroupByCountSQL("type", false)
	builder = r.BuildFilterQuery(builder, flt)
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get type query: %w", err)
	}

	results := make(map[asset.Type]int)
	rows, err := r.client.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get type of assets: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		row := make(map[string]interface{})
		err = rows.MapScan(row)
		if err != nil {
			return nil, err
		}
		typeStr, ok := row["type"].(string)
		if !ok {
			return nil, err
		}
		typeCount, ok := row["count"].(int64)
		if !ok {
			return nil, err
		}
		typeName := asset.Type(typeStr)
		if typeName.IsValid() {
			results[typeName] = int(typeCount)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate over asset types result: %w", err)
	}

	return results, nil
}

// GetCount retrieves number of assets for every type
func (r *AssetRepository) GetCount(ctx context.Context, flt asset.Filter) (int, error) {
	builder := sq.Select("count(1)").
		Where(sq.Eq{"is_deleted": flt.IsDeleted}).
		From("assets")
	builder = r.BuildFilterQuery(builder, flt)
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var total int
	if err := r.client.db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, fmt.Errorf("get asset list: %w", err)
	}

	return total, nil
}

// GetCountByQueryExpr retrieves number of assets for every type based on query expr
func (r *AssetRepository) GetCountByQueryExpr(ctx context.Context, queryExpr queryexpr.ExprStr) (uint32, error) {
	query, err := queryexpr.ValidateAndGetQueryFromExpr(queryExpr)
	if err != nil {
		return 0, err
	}

	total, err := r.getCountByQuery(ctx, query)
	if err != nil {
		return 0, err
	}

	return total, err
}

func (r *AssetRepository) getCountByQuery(ctx context.Context, sqlQuery string) (uint32, error) {
	builder := sq.Select("count(1)").
		From("assets").
		Where(sqlQuery)
	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var total uint32
	if err := r.client.db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, fmt.Errorf("get asset list: %w", err)
	}

	return total, nil
}

// GetCountByIsDeletedAndServicesAndUpdatedAt retrieves number of assets based on IsDeleted, Services, and UpdatedAt
func (r *AssetRepository) GetCountByIsDeletedAndServicesAndUpdatedAt(
	ctx context.Context,
	isDeleted bool,
	services []string,
	thresholdTime time.Time,
) (uint32, error) {
	if len(services) == 0 {
		return 0, asset.ErrEmptyServices
	}
	builder := sq.Select("count(1)").
		From("assets").
		Where(sq.Eq{"is_deleted": isDeleted}).
		Where(sq.Lt{"updated_at": thresholdTime})

	if strings.TrimSpace(services[0]) != asset.AllServicesCleanupConfig {
		builder = builder.Where(sq.Eq{"service": services})
	}

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var total uint32
	if err := r.client.db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, fmt.Errorf("get asset list: %w", err)
	}

	return total, nil
}

// GetByID retrieves asset by its ID
func (r *AssetRepository) GetByID(ctx context.Context, id string) (asset.Asset, error) {
	if !isValidUUID(id) {
		return asset.Asset{}, asset.InvalidError{AssetID: id}
	}

	ast, err := r.getWithPredicate(ctx, sq.Eq{"a.id": id})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{AssetID: id}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error getting asset with ID = %q: %w", id, err)
	}

	return ast, nil
}

// GetByIDWithTx retrieves asset by its ID with Transaction
func (r *AssetRepository) GetByIDWithTx(ctx context.Context, tx *sqlx.Tx, id string) (asset.Asset, error) {
	if !isValidUUID(id) {
		return asset.Asset{}, asset.InvalidError{AssetID: id}
	}

	ast, err := r.getWithPredicateWithTx(ctx, tx, sq.Eq{"a.id": id})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{AssetID: id}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error getting asset with ID = %q: %w", id, err)
	}

	return ast, nil
}

func (r *AssetRepository) GetByURN(ctx context.Context, urn string) (asset.Asset, error) {
	ast, err := r.getWithPredicate(ctx, sq.Eq{"a.urn": urn})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{URN: urn}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error getting asset with URN = %q: %w", urn, err)
	}

	return ast, nil
}

func (r *AssetRepository) GetByURNWithTx(ctx context.Context, tx *sqlx.Tx, urn string) (asset.Asset, error) {
	ast, err := r.getWithPredicateWithTx(ctx, tx, sq.Eq{"a.urn": urn})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{URN: urn}
	}
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error getting asset with Tx for URN = %q: %w", urn, err)
	}

	return ast, nil
}

func (r *AssetRepository) getWithPredicate(ctx context.Context, pred sq.Eq) (asset.Asset, error) {
	return r.getWithPredicateWithTx(ctx, nil, pred)
}

func (r *AssetRepository) getWithPredicateWithTx(ctx context.Context, tx *sqlx.Tx, pred sq.Eq) (asset.Asset, error) {
	query, args, err := r.getAssetSQLWithForUpdate().
		Where(pred).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error building query: %w", err)
	}

	var am AssetModel
	if tx == nil {
		err = r.client.db.GetContext(ctx, &am, query, args...)
	} else {
		err = tx.GetContext(ctx, &am, query, args...)
	}
	if err != nil {
		return asset.Asset{}, err
	}

	owners, err := r.getOwnersWithTx(ctx, tx, am.ID)
	if err != nil {
		return asset.Asset{}, err
	}

	return am.toAsset(owners), nil
}

// GetVersionHistory retrieves the versions of an asset
func (r *AssetRepository) GetVersionHistory(ctx context.Context, flt asset.Filter, id string) ([]asset.Asset, error) {
	if !isValidUUID(id) {
		return nil, asset.InvalidError{AssetID: id}
	}

	size := flt.Size
	if size == 0 {
		size = r.defaultGetMaxSize
	}

	query, args, err := r.getAssetVersionSQL().
		Where(sq.Eq{"a.asset_id": id}).
		OrderBy("string_to_array(version, '.')::int[] DESC").
		Limit(uint64(size)).
		Offset(uint64(flt.Offset)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var assetModels []AssetModel
	if err := r.client.db.SelectContext(ctx, &assetModels, query, args...); err != nil {
		return nil, fmt.Errorf("fetch last versions: %w", err)
	}

	if len(assetModels) == 0 {
		return nil, asset.NotFoundError{AssetID: id}
	}

	avs := make([]asset.Asset, 0, len(assetModels))
	for _, am := range assetModels {
		av, err := am.toAssetVersion()
		if err != nil {
			return nil, fmt.Errorf("convert asset model to asset version: %w", err)
		}

		avs = append(avs, av)
	}

	return avs, nil
}

// GetByVersionWithID retrieves the specific asset version
func (r *AssetRepository) GetByVersionWithID(ctx context.Context, id, version string) (asset.Asset, error) {
	if !isValidUUID(id) {
		return asset.Asset{}, asset.InvalidError{AssetID: id}
	}

	ast, err := r.getByVersion(ctx, id, version, r.GetByID, sq.Eq{
		"a.asset_id": id,
		"a.version":  version,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{AssetID: id}
	}
	if err != nil {
		return asset.Asset{}, err
	}

	return ast, nil
}

func (r *AssetRepository) GetByVersionWithURN(ctx context.Context, urn, version string) (asset.Asset, error) {
	ast, err := r.getByVersion(ctx, urn, version, r.GetByURN, sq.Eq{
		"a.urn":     urn,
		"a.version": version,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return asset.Asset{}, asset.NotFoundError{URN: urn}
	}
	if err != nil {
		return asset.Asset{}, err
	}

	return ast, nil
}

type getAssetFunc func(context.Context, string) (asset.Asset, error)

func (r *AssetRepository) getByVersion(
	ctx context.Context, id, version string, get getAssetFunc, pred sq.Eq,
) (asset.Asset, error) {
	latest, err := get(ctx, id)
	if err != nil {
		return asset.Asset{}, err
	}

	if latest.Version == version {
		return latest, nil
	}

	var ast AssetModel
	query, args, err := r.getAssetVersionSQL().
		Where(pred).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error building query: %w", err)
	}

	err = r.client.db.GetContext(ctx, &ast, query, args...)
	if err != nil {
		return asset.Asset{}, fmt.Errorf("failed fetching asset version: %w", err)
	}

	return ast.toVersionedAsset(latest)
}

// Upsert creates a new asset if it does not exist yet.
// It updates if asset does exist.
// Checking existence is done using "urn", "type", "name", "data", and "service" fields.
func (r *AssetRepository) Upsert(ctx context.Context, ast *asset.Asset) (upsertedAsset *asset.Asset, err error) {
	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) (err error) {
		fetchedAsset, err := r.GetByURNWithTx(ctx, tx, ast.URN)
		if errors.As(err, new(asset.NotFoundError)) {
			// insert flow
			upsertedAsset, err = r.insert(ctx, tx, ast)
			if err != nil {
				return fmt.Errorf("error inserting asset to DB: %w", err)
			}

			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting asset by URN: %w", err)
		}

		// reset IsDeleted flag if asset is resync'd
		ast.IsDeleted = false

		changelog, err := fetchedAsset.Diff(ast)
		if err != nil {
			return fmt.Errorf("error diffing two assets: %w", err)
		}

		ast.ID = fetchedAsset.ID
		upsertedAsset, err = r.update(ctx, tx, ast, &fetchedAsset, changelog)
		if err != nil {
			return fmt.Errorf("error updating asset to DB: %w", err)
		}

		return nil
	})
	if err != nil {
		return upsertedAsset, err
	}

	return upsertedAsset, nil
}

// UpsertPatch creates a new asset if it does not exist yet.
// It updates if asset does exist.
// Checking existence is done using "urn", "type", and "service" fields
// And will revalidate again with additional: "data" and "name" fields.
func (r *AssetRepository) UpsertPatch( //nolint:gocognit
	ctx context.Context,
	ast *asset.Asset,
	patchData map[string]interface{},
) (upsertedAsset *asset.Asset, err error) {
	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) (err error) {
		fetchedAsset, err := r.GetByURNWithTx(ctx, tx, ast.URN)
		if errors.As(err, new(asset.NotFoundError)) {
			// insert flow
			if err := r.validateAsset(*ast); err != nil {
				return err
			}
			upsertedAsset, err = r.insert(ctx, tx, ast)
			if err != nil {
				return fmt.Errorf("error inserting asset to DB: %w", err)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting asset by URN: %w", err)
		}

		// update flow
		var newAsset asset.Asset
		if err := copier.CopyWithOption(&newAsset, fetchedAsset, copier.Option{DeepCopy: true}); err != nil {
			return err
		}
		newAsset.Patch(patchData)
		newAsset.RefreshedAt = ast.RefreshedAt

		// reset IsDeleted flag if asset is resync'd
		newAsset.IsDeleted = false

		if err := r.validateAsset(newAsset); err != nil {
			return err
		}
		changelog, err := fetchedAsset.Diff(&newAsset)
		if err != nil {
			return fmt.Errorf("error diffing two assets: %w", err)
		}

		upsertedAsset, err = r.update(ctx, tx, &newAsset, &fetchedAsset, changelog)
		if err != nil {
			return fmt.Errorf("error updating asset to DB: %w", err)
		}

		return nil
	})
	if err != nil {
		return upsertedAsset, err
	}

	return upsertedAsset, nil
}

func (*AssetRepository) validateAsset(ast asset.Asset) error {
	if ast.URN == "" {
		return fmt.Errorf("urn is required")
	}
	if ast.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !ast.Type.IsValid() {
		return fmt.Errorf("type is invalid")
	}
	if ast.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ast.Data == nil {
		return fmt.Errorf("data is required")
	}
	if ast.Service == "" {
		return fmt.Errorf("service is required")
	}

	return nil
}

// DeleteByID hard delete asset using its ID
func (r *AssetRepository) DeleteByID(ctx context.Context, id string) (urn string, err error) {
	if !isValidUUID(id) {
		return "", asset.InvalidError{AssetID: id}
	}

	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		fetchedAsset, err := r.GetByIDWithTx(ctx, tx, id)
		if err != nil {
			return asset.NotFoundError{AssetID: id}
		}
		urn = fetchedAsset.URN

		err = r.deleteWithPredicate(ctx, tx, sq.Eq{"id": id})
		if err != nil {
			return fmt.Errorf("error deleting asset with ID = %q: %w", id, err)
		}

		return nil
	})
	if err != nil {
		return urn, err
	}

	return urn, nil
}

// DeleteByURN hard delete asset using its URN
func (r *AssetRepository) DeleteByURN(ctx context.Context, urn string) error {
	err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) (err error) {
		err = r.deleteWithPredicate(ctx, tx, sq.Eq{"urn": urn})
		if err != nil {
			return fmt.Errorf("error deleting asset with URN = %q: %w", urn, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *AssetRepository) SoftDeleteByID(
	ctx context.Context,
	executedAt time.Time,
	id, updatedByID string,
) (urn, newVersion string, err error) {
	if !isValidUUID(id) {
		return "", "", asset.InvalidError{AssetID: id}
	}

	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		fetchedAsset, err := r.GetByIDWithTx(ctx, tx, id)
		if err != nil {
			return asset.NotFoundError{AssetID: id}
		}
		urn = fetchedAsset.URN

		if fetchedAsset.IsDeleted {
			return asset.ErrAssetAlreadyDeleted
		}

		newVersion, err = asset.IncreaseMinorVersion(fetchedAsset.Version)
		if err != nil {
			return err
		}

		// just need soft copy since fetchedAsset is not used anymore
		newAsset := fetchedAsset
		newAsset.Version = newVersion
		newAsset.UpdatedAt = executedAt
		newAsset.RefreshedAt = &executedAt
		newAsset.IsDeleted = true
		newAsset.UpdatedBy = user.User{ID: updatedByID}
		newAsset.Changelog = diff.Changelog{
			{
				Type: "delete",
				Path: []string{"is_deleted"},
				From: false,
				To:   true,
			},
		}

		err = r.softDeleteWithPredicate(ctx, tx, sq.Eq{"id": id}, newAsset)
		if err != nil {
			return fmt.Errorf("error deleting asset with ID = %q: %w", id, err)
		}

		return nil
	})
	if err != nil {
		return urn, newVersion, err
	}

	return urn, newVersion, nil
}

func (r *AssetRepository) SoftDeleteByURN(ctx context.Context, executedAt time.Time, urn, updatedByID string) (newVersion string, err error) {
	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) (err error) {
		fetchedAsset, err := r.GetByURNWithTx(ctx, tx, urn)
		if err != nil {
			return asset.NotFoundError{URN: urn}
		}

		if fetchedAsset.IsDeleted {
			return asset.ErrAssetAlreadyDeleted
		}

		newVersion, err = asset.IncreaseMinorVersion(fetchedAsset.Version)
		if err != nil {
			return err
		}

		// just need soft copy since fetchedAsset is not used anymore
		newAsset := fetchedAsset
		newAsset.Version = newVersion
		newAsset.UpdatedAt = executedAt
		newAsset.RefreshedAt = &executedAt
		newAsset.IsDeleted = true
		newAsset.UpdatedBy = user.User{ID: updatedByID}
		newAsset.Changelog = diff.Changelog{
			{
				Type: "delete",
				Path: []string{"is_deleted"},
				From: false,
				To:   true,
			},
		}

		err = r.softDeleteWithPredicate(ctx, tx, sq.Eq{"urn": urn}, newAsset)
		if err != nil {
			return fmt.Errorf("error deleting asset with URN = %q: %w", urn, err)
		}

		return nil
	})
	if err != nil {
		return newVersion, err
	}

	return newVersion, nil
}

func (r *AssetRepository) DeleteByIsDeletedAndServicesAndUpdatedAt(
	ctx context.Context,
	isDeleted bool,
	services []string,
	thresholdTime time.Time,
) (urns []string, err error) {
	if len(services) == 0 {
		return nil, asset.ErrEmptyServices
	}
	err = r.client.RunWithinTx(ctx, func(*sqlx.Tx) error {
		builder := sq.Delete("assets").
			Where(sq.Eq{"is_deleted": isDeleted}).
			Where(sq.Lt{"updated_at": thresholdTime}).
			Suffix("RETURNING urn")

		if strings.TrimSpace(services[0]) != asset.AllServicesCleanupConfig {
			builder = builder.Where(sq.Eq{"service": services})
		}

		query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return fmt.Errorf("error building query: %w", err)
		}

		return r.client.db.SelectContext(ctx, &urns, query, args...)
	})

	return urns, err
}

func (r *AssetRepository) DeleteByQueryExpr(ctx context.Context, queryExpr queryexpr.ExprStr) ([]string, error) {
	var urns []string
	err := r.client.RunWithinTx(ctx, func(*sqlx.Tx) error {
		query, err := queryexpr.ValidateAndGetQueryFromExpr(queryExpr)
		if err != nil {
			return err
		}

		urns, err = r.deleteByQueryAndReturnURNS(ctx, query)

		return err
	})

	return urns, err
}

func (r *AssetRepository) SoftDeleteByQueryExpr(
	ctx context.Context,
	executedAt time.Time,
	updatedByID string,
	queryExpr queryexpr.ExprStr,
) (updatedAssets []asset.Asset, err error) {
	err = r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
		query, err := queryexpr.ValidateAndGetQueryFromExpr(queryExpr)
		if err != nil {
			return err
		}

		updatedAssets, err = r.softDeleteByQuery(ctx, tx, query, executedAt, updatedByID)
		if err != nil {
			return fmt.Errorf("error soft deleting assets by query: %w", err)
		}

		softDeleteChangelog := diff.Changelog{
			{
				Type: "delete",
				Path: []string{"is_deleted"},
				From: false,
				To:   true,
			},
		}
		return r.insertAssetVersions(ctx, tx, updatedAssets, softDeleteChangelog)
	})

	return updatedAssets, err
}

// deleteByQueryAndReturnURNS remove all assets that match to query and return array of urn of asset that deleted.
func (r *AssetRepository) deleteByQueryAndReturnURNS(ctx context.Context, whereCondition string) ([]string, error) {
	builder := sq.Delete("assets").
		Where(whereCondition).
		Suffix("RETURNING urn")

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var urns []string
	if err := r.client.db.SelectContext(ctx, &urns, query, args...); err != nil {
		return nil, err
	}

	return urns, nil
}

// softDeleteByQuery soft delete all assets that match to query
func (r *AssetRepository) softDeleteByQuery(
	ctx context.Context,
	tx *sqlx.Tx,
	whereCondition string,
	executedAt time.Time,
	updatedByID string,
) ([]asset.Asset, error) {
	updateCTE := sq.Update("assets").
		Set("is_deleted", true).
		Set("updated_at", executedAt).
		Set("refreshed_at", executedAt).
		Set("updated_by", updatedByID).
		Where(whereCondition).
		Suffix("RETURNING *").
		Prefix("WITH assets AS (").
		Suffix(")")

	returnQuery := r.getAssetSQLWithIsDeleted(true, false)

	fullQuery := returnQuery.PrefixExpr(updateCTE)
	query, args, err := fullQuery.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var ams []AssetModel
	err = sqlx.SelectContext(ctx, tx, &ams, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	assets := make([]asset.Asset, 0, len(ams))
	for _, am := range ams {
		owners, err := r.getOwnersWithTx(ctx, tx, am.ID)
		if err != nil {
			return nil, fmt.Errorf("get owners with ID: %s, %w", am.ID, err)
		}

		assets = append(assets, am.toAsset(owners))
	}

	return assets, nil
}

func (r *AssetRepository) AddProbe(ctx context.Context, assetURN string, probe *asset.Probe) error {
	probe.AssetURN = assetURN
	probe.CreatedAt = time.Now().UTC()
	if probe.Timestamp.IsZero() {
		probe.Timestamp = probe.CreatedAt
	} else {
		probe.Timestamp = probe.Timestamp.UTC()
	}

	insert := sq.Insert("asset_probes")
	if probe.ID != "" {
		insert = insert.Columns("id", "asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at").
			Values(probe.ID, assetURN, probe.Status, probe.StatusReason, probe.Metadata, probe.Timestamp, probe.CreatedAt)
	} else {
		insert = insert.Columns("asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at").
			Values(assetURN, probe.Status, probe.StatusReason, probe.Metadata, probe.Timestamp, probe.CreatedAt)
	}

	query, args, err := insert.Suffix("RETURNING \"id\"").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert asset probe query: %w", err)
	}

	if err = r.client.db.QueryRowContext(ctx, query, args...).Scan(&probe.ID); err != nil {
		switch e := checkPostgresError(err); {
		case errors.Is(e, errForeignKeyViolation):
			return asset.NotFoundError{URN: assetURN}

		case errors.Is(e, errDuplicateKey):
			return asset.ErrProbeExists
		}

		return fmt.Errorf("run insert asset probe query: %w", err)
	}

	return nil
}

func (r *AssetRepository) GetProbes(ctx context.Context, assetURN string) ([]asset.Probe, error) {
	query, args, err := sq.Select(
		"id", "asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at",
	).From("asset_probes").
		OrderBy("timestamp").
		Where(sq.Eq{"asset_urn": assetURN}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building get asset probes query: %w", err)
	}

	var models []AssetProbeModel
	if err := r.client.db.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, fmt.Errorf("error running get asset probes query: %w", err)
	}

	results := []asset.Probe{}
	for _, m := range models {
		results = append(results, m.toAssetProbe())
	}

	return results, nil
}

func (r *AssetRepository) GetProbesWithFilter(ctx context.Context, flt asset.ProbesFilter) (map[string][]asset.Probe, error) {
	stmt := sq.Select(
		"id", "asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at",
	).From("asset_probes").
		OrderBy("asset_urn", "timestamp DESC")

	if len(flt.AssetURNs) > 0 {
		stmt = stmt.Where(sq.Eq{"asset_urn": flt.AssetURNs})
	}
	if !flt.NewerThan.IsZero() {
		stmt = stmt.Where(sq.GtOrEq{"timestamp": flt.NewerThan})
	}
	if !flt.OlderThan.IsZero() {
		stmt = stmt.Where(sq.LtOrEq{"timestamp": flt.OlderThan})
	}
	if flt.MaxRows > 0 {
		stmt = stmt.Column("RANK() OVER (PARTITION BY asset_urn ORDER BY timestamp desc) rank_number")
		stmt = sq.Select(
			"id", "asset_urn", "status", "status_reason", "metadata", "timestamp", "created_at",
		).FromSelect(stmt, "ap").
			Where(sq.LtOrEq{"rank_number": flt.MaxRows})
	}

	query, args, err := stmt.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("get probes with filter: build query: %w", err)
	}

	var probes []AssetProbeModel
	if err := r.client.db.SelectContext(ctx, &probes, query, args...); err != nil {
		return nil, fmt.Errorf("error running get asset probes query: %w", err)
	}

	results := make(map[string][]asset.Probe, len(probes))
	for _, p := range probes {
		results[p.AssetURN] = append(results[p.AssetURN], p.toAssetProbe())
	}

	return results, nil
}

func (r *AssetRepository) deleteWithPredicate(ctx context.Context, tx *sqlx.Tx, pred sq.Eq) error {
	query, args, err := sq.Delete("assets").
		Where(pred).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	return r.execContext(ctx, tx, query, args...)
}

func (r *AssetRepository) softDeleteWithPredicate(ctx context.Context, tx *sqlx.Tx, pred sq.Eq, newAsset asset.Asset) error {
	query, args, err := sq.Update("assets").
		Set("is_deleted", true).
		Set("updated_at", newAsset.UpdatedAt).
		Set("refreshed_at", newAsset.RefreshedAt).
		Set("updated_by", newAsset.UpdatedBy.ID).
		Set("version", newAsset.Version).
		Where(pred).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}
	if err := r.execContext(ctx, tx, query, args...); err != nil {
		return err
	}

	return r.insertAssetVersion(ctx, tx, &newAsset, newAsset.Changelog)
}

func (r *AssetRepository) insert(ctx context.Context, tx *sqlx.Tx, ast *asset.Asset) (*asset.Asset, error) {
	currentTime := time.Now()
	if ast.RefreshedAt != nil {
		currentTime = *ast.RefreshedAt
	}
	ast.CreatedAt = currentTime
	ast.UpdatedAt = currentTime
	insertCTE := sq.Insert("assets").
		Columns("urn", "type", "service", "name", "description", "data", "url", "labels",
			"created_at", "updated_by", "updated_at", "refreshed_at", "version", "is_deleted").
		Values(ast.URN, ast.Type, ast.Service, ast.Name, ast.Description, ast.Data, ast.URL, ast.Labels,
			ast.CreatedAt, ast.UpdatedBy.ID, ast.UpdatedAt, currentTime, asset.BaseVersion, ast.IsDeleted).
		Suffix("RETURNING *").
		Prefix("WITH assets AS (").
		Suffix(")")
	returnQuery := r.getAssetSQL()

	// Combine CTE and main query
	fullQuery := returnQuery.PrefixExpr(insertCTE)

	query, args, err := fullQuery.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert query: %w", err)
	}

	var am AssetModel
	err = sqlx.GetContext(ctx, tx, &am, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "assets_idx_urn") {
			return nil, asset.ErrURNExist
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	users, err := r.createOrFetchUsers(ctx, tx, ast.Owners)
	if err != nil {
		return nil, fmt.Errorf("create and fetch owners: %w", err)
	}
	err = r.insertOwners(ctx, tx, am.ID, users)
	if err != nil {
		return nil, fmt.Errorf("run insert owners query: %w", err)
	}

	insertedAsset := am.toAsset(users)

	if err := r.insertAssetVersion(ctx, tx, &insertedAsset, diff.Changelog{}); err != nil {
		return nil, err
	}

	return &insertedAsset, nil
}

func (r *AssetRepository) update(ctx context.Context, tx *sqlx.Tx, newAsset, oldAsset *asset.Asset, clog diff.Changelog) (*asset.Asset, error) {
	assetID := oldAsset.ID
	if !isValidUUID(assetID) {
		return nil, asset.InvalidError{AssetID: assetID}
	}

	currentTime := time.Now()
	// for Upsert API calls, to make currentTime value same for both Postgresql and Elasticsearch,
	// the currentTime already filled in UpsertPatchAssetWithoutLineage/UpsertAssetWithoutLineage
	if newAsset.RefreshedAt != nil {
		currentTime = *newAsset.RefreshedAt
	}

	if len(clog) == 0 {
		if newAsset.RefreshedAt == nil || newAsset.RefreshedAt == oldAsset.RefreshedAt {
			return newAsset, nil
		}

		updatedAsset, err := r.updateAssetRefreshedAt(ctx, tx, assetID, currentTime)
		if err != nil {
			return nil, fmt.Errorf("error updating asset's refreshed_at: %w", err)
		}
		return &updatedAsset, err
	}

	// managing owners
	newAssetOwners, err := r.createOrFetchUsers(ctx, tx, newAsset.Owners)
	if err != nil {
		return nil, fmt.Errorf("error creating and fetching owners: %w", err)
	}
	toInserts, toRemoves := r.compareOwners(oldAsset.Owners, newAssetOwners)
	if err := r.insertOwners(ctx, tx, assetID, toInserts); err != nil {
		return nil, fmt.Errorf("error inserting asset's new owners: %w", err)
	}
	if err := r.removeOwners(ctx, tx, assetID, toRemoves); err != nil {
		return nil, fmt.Errorf("error removing asset's old owners: %w", err)
	}

	// update assets
	newVersion, err := asset.IncreaseMinorVersion(oldAsset.Version)
	if err != nil {
		return nil, err
	}
	newAsset.Version = newVersion
	newAsset.ID = assetID
	newAsset.UpdatedAt = currentTime
	newAsset.RefreshedAt = &currentTime

	updatedAsset, err := r.updateAsset(ctx, tx, assetID, newAsset)
	if err != nil {
		return nil, err
	}

	// insert versions
	if err := r.insertAssetVersion(ctx, tx, newAsset, clog); err != nil {
		return nil, err
	}

	return &updatedAsset, nil
}

func (r *AssetRepository) updateAsset(ctx context.Context, tx *sqlx.Tx, assetID string, newAsset *asset.Asset) (asset.Asset, error) {
	updateCTE := sq.Update("assets").
		Set("urn", newAsset.URN).
		Set("type", newAsset.Type).
		Set("service", newAsset.Service).
		Set("name", newAsset.Name).
		Set("description", newAsset.Description).
		Set("data", newAsset.Data).
		Set("url", newAsset.URL).
		Set("labels", newAsset.Labels).
		Set("is_deleted", newAsset.IsDeleted).
		Set("updated_at", newAsset.UpdatedAt).
		Set("refreshed_at", *newAsset.RefreshedAt).
		Set("updated_by", newAsset.UpdatedBy.ID).
		Set("version", newAsset.Version).
		Where(sq.Eq{"id": assetID}).
		Suffix("RETURNING *").
		Prefix("WITH assets AS (").
		Suffix(")")
	returnQuery := r.getAssetSQL()

	fullQuery := returnQuery.PrefixExpr(updateCTE)
	query, args, err := fullQuery.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return asset.Asset{}, fmt.Errorf("build query: %w", err)
	}

	var am AssetModel
	err = sqlx.GetContext(ctx, tx, &am, query, args...)
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error executing query: %w", err)
	}

	owners, err := r.getOwnersWithTx(ctx, tx, am.ID)
	if err != nil {
		return asset.Asset{}, err
	}

	return am.toAsset(owners), nil
}

func (r *AssetRepository) updateAssetRefreshedAt(ctx context.Context, tx *sqlx.Tx, assetID string, currentTime time.Time) (asset.Asset, error) {
	updateCTE := sq.Update("assets").
		Set("refreshed_at", currentTime).
		Where(sq.Eq{"id": assetID}).
		Suffix("RETURNING *").
		Prefix("WITH assets AS (").
		Suffix(")")
	returnQuery := r.getAssetSQL()

	fullQuery := returnQuery.PrefixExpr(updateCTE)
	query, args, err := fullQuery.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return asset.Asset{}, fmt.Errorf("build query: %w", err)
	}

	var am AssetModel
	err = sqlx.GetContext(ctx, tx, &am, query, args...)
	if err != nil {
		return asset.Asset{}, fmt.Errorf("error executing query: %w", err)
	}

	owners, err := r.getOwnersWithTx(ctx, tx, am.ID)
	if err != nil {
		return asset.Asset{}, err
	}

	return am.toAsset(owners), nil
}

// insertAssetVersion implemented after insert and update, so newest assert version is same with current asset
func (r *AssetRepository) insertAssetVersion(ctx context.Context, execer sqlx.ExecerContext, newAsset *asset.Asset, clog diff.Changelog) error {
	if newAsset == nil {
		return asset.ErrNilAsset
	}

	if clog == nil {
		return fmt.Errorf("changelog is nil when insert to asset version")
	}

	if newAsset.UpdatedBy.ID == "" {
		return fmt.Errorf("user not found with ID: %s", newAsset.UpdatedBy.ID)
	}

	jsonChangelog, err := json.Marshal(clog)
	if err != nil {
		return err
	}
	query, args, err := sq.Insert("assets_versions").
		Columns("asset_id", "urn", "type", "service",
			"name", "description", "data", "labels",
			"created_at", "updated_at", "updated_by", "version",
			"owners", "is_deleted", "changelog").
		Values(newAsset.ID, newAsset.URN, newAsset.Type, newAsset.Service,
			newAsset.Name, newAsset.Description, newAsset.Data, newAsset.Labels,
			newAsset.CreatedAt, newAsset.UpdatedAt, newAsset.UpdatedBy.ID, newAsset.Version,
			newAsset.Owners, newAsset.IsDeleted, jsonChangelog).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if err = r.execContext(ctx, execer, query, args...); err != nil {
		return fmt.Errorf("insert asset version: %w", err)
	}

	return nil
}

// insertAssetVersions run same as insertAssetVersion, but for bulk insert with same changelogs
func (r *AssetRepository) insertAssetVersions(
	ctx context.Context,
	execer sqlx.ExecerContext,
	newAssets []asset.Asset,
	clog diff.Changelog,
) error {
	if len(newAssets) == 0 {
		return nil // nothing to insert
	}

	builder := sq.Insert("assets_versions").
		Columns("asset_id", "urn", "type", "service", "name", "description", "data", "labels",
			"created_at", "updated_at", "updated_by", "version", "owners", "is_deleted", "changelog").
		PlaceholderFormat(sq.Dollar)

	for _, newAsset := range newAssets {
		if newAsset.ID == "" {
			return asset.ErrNilAsset
		}

		builder = builder.Values(
			newAsset.ID,
			newAsset.URN,
			newAsset.Type,
			newAsset.Service,
			newAsset.Name,
			newAsset.Description,
			newAsset.Data,
			newAsset.Labels,
			newAsset.CreatedAt,
			newAsset.UpdatedAt,
			newAsset.UpdatedBy.ID,
			newAsset.Version,
			newAsset.Owners,
			newAsset.IsDeleted,
			clog,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build bulk insert query: %w", err)
	}

	if err := r.execContext(ctx, execer, query, args...); err != nil {
		return fmt.Errorf("bulk insert asset versions: %w", err)
	}

	return nil
}

func (r *AssetRepository) getOwners(ctx context.Context, assetID string) ([]user.User, error) {
	return r.getOwnersWithTx(ctx, nil, assetID)
}

func (r *AssetRepository) getOwnersWithTx(ctx context.Context, tx *sqlx.Tx, assetID string) ([]user.User, error) {
	if !isValidUUID(assetID) {
		return nil, asset.InvalidError{AssetID: assetID}
	}

	var userModels UserModels

	query := `
		SELECT
			u.id as "id",
			u.email as "email",
			u.provider as "provider"
		FROM asset_owners ao
		JOIN users u on ao.user_id = u.id
		WHERE asset_id = $1`

	var err error
	if tx == nil {
		err = r.client.db.SelectContext(ctx, &userModels, query, assetID)
	} else {
		err = tx.SelectContext(ctx, &userModels, query, assetID)
	}
	if err != nil {
		return nil, fmt.Errorf("get asset owners: %w", err)
	}

	return userModels.toUsers(), nil
}

// insertOwners inserts relation of asset id and user id
func (r *AssetRepository) insertOwners(ctx context.Context, execer sqlx.ExecerContext, assetID string, owners []user.User) error {
	if len(owners) == 0 {
		return nil
	}

	if !isValidUUID(assetID) {
		return asset.InvalidError{AssetID: assetID}
	}

	sqlb := sq.Insert("asset_owners").
		Columns("asset_id", "user_id")
	for _, o := range owners {
		sqlb = sqlb.Values(assetID, o.ID)
	}

	qry, args, err := sqlb.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build insert owners SQL: %w", err)
	}

	if err := r.execContext(ctx, execer, qry, args...); err != nil {
		return fmt.Errorf("error running insert owners query: %w", err)
	}

	return nil
}

func (r *AssetRepository) removeOwners(ctx context.Context, execer sqlx.ExecerContext, assetID string, owners []user.User) error {
	if len(owners) == 0 {
		return nil
	}

	if !isValidUUID(assetID) {
		return asset.InvalidError{AssetID: assetID}
	}

	var userIDs []string
	args := []interface{}{assetID}
	for i, owner := range owners {
		userIDs = append(userIDs, fmt.Sprintf("$%d", i+2))
		args = append(args, owner.ID)
	}
	query := fmt.Sprintf(
		`DELETE FROM asset_owners WHERE asset_id = $1 AND user_id in (%s)`,
		strings.Join(userIDs, ","),
	)
	if err := r.execContext(ctx, execer, query, args...); err != nil {
		return fmt.Errorf("run delete owners query: %w", err)
	}

	return nil
}

func (r *AssetRepository) compareOwners(current, newOwners []user.User) (toInserts, toRemove []user.User) {
	if len(current) == 0 && len(newOwners) == 0 {
		return nil, nil
	}

	currMap := map[string]int{}
	for _, curr := range current {
		currMap[curr.ID] = 1
	}

	for _, n := range newOwners {
		_, exists := currMap[n.ID]
		if exists {
			// if exists, it means that both new and current have it.
			// we remove it from the map,
			// so that what's left in the map is the that only exists in current
			// and have to be removed
			delete(currMap, n.ID)
		} else {
			toInserts = append(toInserts, user.User{ID: n.ID})
		}
	}

	for id := range currMap {
		toRemove = append(toRemove, user.User{ID: id})
	}

	return toInserts, toRemove
}

func (r *AssetRepository) createOrFetchUsers(ctx context.Context, tx *sqlx.Tx, users []user.User) ([]user.User, error) {
	ids := make(map[string]struct{}, len(users))
	var results []user.User
	for _, u := range users {
		u := u
		if u.ID != "" {
			if _, ok := ids[u.ID]; ok {
				continue
			}
			ids[u.ID] = struct{}{}
			results = append(results, u)
			continue
		}

		var (
			userID      string
			fetchedUser user.User
			err         error
		)
		fetchedUser, err = r.userRepo.GetByEmailWithTx(ctx, tx, u.Email)

		switch {
		case errors.As(err, &user.NotFoundError{}):
			u.Provider = r.defaultUserProvider
			userID, err = r.userRepo.CreateWithTx(ctx, tx, &u)
			if err != nil {
				return nil, fmt.Errorf("create owner: %w", err)
			}

		case err != nil:
			return nil, fmt.Errorf("get owner's ID: %w", err)

		default: // case err == nil:
			userID = fetchedUser.ID
		}

		if _, ok := ids[userID]; ok {
			continue
		}
		ids[userID] = struct{}{}
		u.ID = userID
		results = append(results, u)
	}

	return results, nil
}

func (*AssetRepository) execContext(ctx context.Context, execer sqlx.ExecerContext, query string, args ...interface{}) error {
	res, err := execer.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error running query: %w", err)
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		return errors.New("query affected 0 rows")
	}

	return nil
}

func (*AssetRepository) getAssetsGroupByCountSQL(columnName string, isDeleted bool) sq.SelectBuilder {
	return sq.Select(columnName, "count(1)").
		From("assets").
		Where(sq.Eq{"is_deleted": isDeleted}).
		GroupBy(columnName)
}

func (*AssetRepository) getAssetSQL() sq.SelectBuilder {
	selectAssetQuery := sq.Select(`
		a.id as id,
		a.urn as urn,
		a.type as type,
		a.name as name,
		a.service as service,
		a.description as description,
		a.data as data,
		COALESCE(a.url, '') as url,
		a.labels as labels,
		a.is_deleted as is_deleted,
		a.version as version,
		a.created_at as created_at,
		a.updated_at as updated_at,
		a.refreshed_at as refreshed_at,
		u.id as "updated_by.id",
		u.email as "updated_by.email",
		u.provider as "updated_by.provider",
		u.created_at as "updated_by.created_at",
		u.updated_at as "updated_by.updated_at"
		`).
		From("assets a").
		LeftJoin("users u ON a.updated_by = u.id")

	return selectAssetQuery
}

func (r *AssetRepository) getAssetSQLWithIsDeleted(isDeleted, doForUpdate bool) sq.SelectBuilder {
	query := r.getAssetSQL().
		Where(sq.Eq{"a.is_deleted": isDeleted})

	if doForUpdate {
		return query.Suffix("FOR UPDATE OF a")
	}

	return query
}

func (r *AssetRepository) getAssetSQLWithForUpdate() sq.SelectBuilder {
	return r.getAssetSQL().
		Suffix("FOR UPDATE OF a")
}

func (*AssetRepository) getAssetVersionSQL() sq.SelectBuilder {
	return sq.Select(`
		a.asset_id as id,
		a.urn as urn,
		a.type as type,
		a.name as name,
		a.service as service,
		a.description as description,
		a.data as data,
		a.labels as labels,
		a.is_deleted as is_deleted,
		a.version as version,
		a.created_at as created_at,
		a.updated_at as updated_at,
		a.changelog as changelog,
		a.owners as owners,
		u.id as "updated_by.id",
		u.email as "updated_by.email",
		u.provider as "updated_by.provider",
		u.created_at as "updated_by.created_at",
		u.updated_at as "updated_by.updated_at"
		`).
		From("assets_versions a").
		LeftJoin("users u ON a.updated_by = u.id")
}

// BuildFilterQuery retrieves the sql query based on applied filter in the queryString
func (r *AssetRepository) BuildFilterQuery(builder sq.SelectBuilder, flt asset.Filter) sq.SelectBuilder {
	if len(flt.Types) > 0 {
		builder = builder.Where(sq.Eq{"type": flt.Types})
	}

	if len(flt.Services) > 0 {
		builder = builder.Where(sq.Eq{"service": flt.Services})
	}

	if len(flt.QueryFields) > 0 && flt.Query != "" {
		orClause := sq.Or{}

		for _, field := range flt.QueryFields {
			finalQuery := field

			if strings.Contains(field, "data") {
				finalQuery = r.buildDataField(
					strings.TrimPrefix(field, "data."),
					false,
				)
			}
			orClause = append(orClause, sq.ILike{
				finalQuery: fmt.Sprint("%", flt.Query, "%"),
			})
		}
		builder = builder.Where(orClause)
	}

	if len(flt.Data) > 0 {
		for key, vals := range flt.Data {
			if len(vals) == 1 && vals[0] == "_nonempty" {
				field := r.buildDataField(key, true)
				whereClause := sq.And{
					sq.NotEq{field: nil},    // IS NOT NULL (field exists)
					sq.NotEq{field: "null"}, // field is not "null" JSON
					sq.NotEq{field: "[]"},   // field is not empty array
					sq.NotEq{field: "{}"},   // field is not empty object
					sq.NotEq{field: "\"\""}, // field is not empty string
				}
				builder = builder.Where(whereClause)
			} else {
				dataOrClause := sq.Or{}
				for _, v := range vals {
					finalQuery := r.buildDataField(key, false)
					dataOrClause = append(dataOrClause, sq.Eq{finalQuery: v})
				}

				builder = builder.Where(dataOrClause)
			}
		}
	}

	return builder
}

// buildFilterQuery retrieves the ordered sql query based on the sorting filter used in queryString
func (r *AssetRepository) buildOrderQuery(builder sq.SelectBuilder, flt asset.Filter) sq.SelectBuilder {
	if flt.SortBy == "" {
		return builder
	}

	orderDirection := "ASC"
	if flt.SortDirection != "" {
		orderDirection = flt.SortDirection
	}

	return builder.OrderBy(flt.SortBy + " " + orderDirection)
}

// buildDataField is a helper function to build nested data fields
func (r *AssetRepository) buildDataField(key string, asJsonB bool) (finalQuery string) {
	var queries []string

	queries = append(queries, "data")
	nestedParams := strings.Split(key, ".")
	totalParams := len(nestedParams)
	for i := 0; i < totalParams-1; i++ {
		nestedQuery := fmt.Sprintf("->'%s'", nestedParams[i])
		queries = append(queries, nestedQuery)
	}

	var lastParam string
	if asJsonB {
		lastParam = fmt.Sprintf("->'%s'", nestedParams[totalParams-1])
	} else {
		lastParam = fmt.Sprintf("->>'%s'", nestedParams[totalParams-1])
	}

	queries = append(queries, lastParam)
	finalQuery = strings.Join(queries, "")

	return finalQuery
}

// NewAssetRepository initializes user repository clients
func NewAssetRepository(c *Client, userRepo *UserRepository, defaultGetMaxSize int, defaultUserProvider string) (*AssetRepository, error) {
	if c == nil {
		return nil, errors.New("postgres client is nil")
	}
	if defaultGetMaxSize == 0 {
		defaultGetMaxSize = DEFAULT_MAX_RESULT_SIZE
	}
	if defaultUserProvider == "" {
		defaultUserProvider = "unknown"
	}

	return &AssetRepository{
		client:              c,
		defaultGetMaxSize:   defaultGetMaxSize,
		defaultUserProvider: defaultUserProvider,
		userRepo:            userRepo,
	}, nil
}
