package elasticsearch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	store "github.com/goto/compass/internal/store/elasticsearch"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/goto/salt/log"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchByUrnFields struct {
	UpdatedAt   time.Time `json:"updated_at"`
	RefreshedAt time.Time `json:"refreshed_at"`
	Version     string    `json:"version"`
	IsDeleted   bool      `json:"is_deleted"`
}

func TestDiscoveryRepository_Upsert(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
	)

	t.Run("should return error if id empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)
		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "",
			Type:    asset.Type("table"),
			Service: bigqueryService,
		})
		assert.ErrorIs(t, err, asset.ErrEmptyID)
	})

	t.Run("should return error if type is not known", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "sample-id",
			Type:    asset.Type("unknown-type"),
			Service: bigqueryService,
		})
		assert.ErrorIs(t, err, asset.ErrUnknownType)
	})

	t.Run("should return error if response.body.Errors is true", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)
		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		// upsert with create_time as a object
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "sample-id",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			Data: map[string]interface{}{
				"create_time": map[string]interface{}{
					"seconds": 1618103237,
					"nanos":   897000000,
				},
			},
		})
		require.NoError(t, err)

		// upsert with create_time as a string
		err = repo.Upsert(ctx, asset.Asset{
			ID:      "sample-id",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			Data: map[string]interface{}{
				"create_time": "2023-04-10T22:33:57.897Z",
			},
		})
		assert.EqualError(t, err, "discovery error: IndexDoc: doc ID 'sample-id': "+
			"index 'bigquery-test': "+
			"elasticsearch code 'mapper_parsing_exception': "+
			"object mapping for [data.create_time] tried to parse field [create_time] as object, but found a concrete value")
	})

	t.Run("should insert asset to the correct index by its service", func(t *testing.T) {
		ast := asset.Asset{
			ID:          "sample-id",
			URN:         "sample-urn",
			Type:        asset.Type("table"),
			Service:     bigqueryService,
			Name:        "sample-name",
			Description: "sample-description",
			Data: map[string]interface{}{
				"foo": map[string]interface{}{
					"company": "gotocompany",
				},
			},
			Labels: map[string]string{
				"bar": "foo",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})
		err = repo.Upsert(ctx, ast)
		assert.NoError(t, err)

		res, err := cli.API.Get(bigqueryService, ast.ID)
		require.NoError(t, err)
		require.False(t, res.IsError())

		var payload struct {
			Source asset.Asset `json:"_source"`
		}
		err = json.NewDecoder(res.Body).Decode(&payload)
		require.NoError(t, err)

		assert.Equal(t, ast.ID, payload.Source.ID)
		assert.Equal(t, ast.URN, payload.Source.URN)
		assert.Equal(t, ast.Type, payload.Source.Type)
		assert.Equal(t, ast.Service, payload.Source.Service)
		assert.Equal(t, ast.Name, payload.Source.Name)
		assert.Equal(t, ast.Description, payload.Source.Description)
		assert.Equal(t, ast.Data, payload.Source.Data)
		assert.Equal(t, ast.Labels, payload.Source.Labels)
		assert.WithinDuration(t, ast.CreatedAt, payload.Source.CreatedAt, 0)
		assert.WithinDuration(t, ast.UpdatedAt, payload.Source.UpdatedAt, 0)
	})

	t.Run("should update existing asset if ID exists", func(t *testing.T) {
		existingAsset := asset.Asset{
			ID:          "existing-id",
			URN:         "existing-urn",
			Type:        asset.Type("table"),
			Service:     bigqueryService,
			Name:        "existing-name",
			Description: "existing-description",
		}
		newAsset := existingAsset
		newAsset.URN = "new-urn"
		newAsset.Name = "new-name"
		newAsset.Description = "new-description"

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)
		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		err = repo.Upsert(ctx, existingAsset)
		assert.NoError(t, err)
		err = repo.Upsert(ctx, newAsset)
		assert.NoError(t, err)

		res, err := cli.API.Get(bigqueryService, existingAsset.ID)
		require.NoError(t, err)
		require.False(t, res.IsError())

		var payload struct {
			Source asset.Asset `json:"_source"`
		}
		err = json.NewDecoder(res.Body).Decode(&payload)
		require.NoError(t, err)

		assert.Equal(t, existingAsset.ID, payload.Source.ID)
		assert.Equal(t, newAsset.URN, payload.Source.URN)
		assert.Equal(t, newAsset.Name, payload.Source.Name)
		assert.Equal(t, newAsset.Description, payload.Source.Description)
	})
}

func TestDiscoveryRepository_DeleteByID(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		kafkaService    = "kafka-test"
	)

	t.Run("should return error if id empty", func(t *testing.T) {
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})
		err = repo.DeleteByID(ctx, "")
		assert.ErrorIs(t, err, asset.ErrEmptyID)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := asset.Asset{
			ID:      "delete-id",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			URN:     "some-urn",
		}

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		err = repo.Upsert(ctx, ast)
		require.NoError(t, err)

		err = repo.DeleteByID(ctx, ast.ID)
		assert.NoError(t, err)

		res, err := cli.Search(
			cli.Search.WithBody(strings.NewReader(`{"query":{"term":{"_id": "delete-id"}}}`)),
			cli.Search.WithIndex("_all"),
		)
		require.NoError(t, err)
		assert.False(t, res.IsError())

		var body struct {
			Hits struct {
				Total elastic.TotalHits `json:"total"`
			} `json:"hits"`
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

		assert.Equal(t, int64(0), body.Hits.Total.Value)
	})

	t.Run("should ignore unavailable indices", func(t *testing.T) {
		ast1 := asset.Asset{
			ID:      "id1",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			URN:     "urn1",
		}
		ast2 := asset.Asset{
			ID:      "id2",
			Type:    asset.Type("topic"),
			Service: kafkaService,
			URN:     "urn2",
		}
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		err = repo.Upsert(ctx, ast1)
		require.NoError(t, err)

		err = repo.Upsert(ctx, ast2)
		require.NoError(t, err)

		_, err = cli.Indices.Close([]string{kafkaService})
		require.NoError(t, err)

		err = repo.DeleteByID(ctx, ast1.ID)
		assert.NoError(t, err)
	})
}

func TestDiscoveryRepository_DeleteByURN(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		kafkaService    = "kafka-test"
	)

	cli, err := esTestServer.NewClient()
	require.NoError(t, err)

	esClient, err := store.NewClient(
		log.NewNoop(), store.Config{}, store.WithClient(cli),
	)
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

	t.Run("should return error if the given urn is empty", func(t *testing.T) {
		err = repo.DeleteByURN(ctx, "")
		assert.ErrorIs(t, err, asset.ErrEmptyURN)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := asset.Asset{
			ID:      "delete-id",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			URN:     "some-urn",
		}

		err = repo.Upsert(ctx, ast)
		require.NoError(t, err)

		err = repo.DeleteByURN(ctx, ast.URN)
		assert.NoError(t, err)

		res, err := cli.Search(
			cli.Search.WithBody(strings.NewReader(`{"query":{"term":{"urn.keyword": "some-urn"}}}`)),
			cli.Search.WithIndex("_all"),
		)
		require.NoError(t, err)
		assert.False(t, res.IsError())

		var body struct {
			Hits struct {
				Total elastic.TotalHits `json:"total"`
			} `json:"hits"`
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
		assert.Equal(t, int64(0), body.Hits.Total.Value)
	})

	t.Run("should ignore unavailable indices", func(t *testing.T) {
		ast1 := asset.Asset{
			ID:      "id1",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			URN:     "urn1",
		}
		ast2 := asset.Asset{
			ID:      "id2",
			Type:    asset.Type("topic"),
			Service: kafkaService,
			URN:     "urn2",
		}
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		err = repo.Upsert(ctx, ast1)
		require.NoError(t, err)

		err = repo.Upsert(ctx, ast2)
		require.NoError(t, err)

		_, err = cli.Indices.Close([]string{kafkaService})
		require.NoError(t, err)

		err = repo.DeleteByURN(ctx, ast1.URN)
		assert.NoError(t, err)
	})
}

func TestDiscoveryRepository_SoftDeleteByURN(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		kafkaService    = "kafka-test"
		currentTime     = time.Now().UTC()
	)

	cli, err := esTestServer.NewClient()
	require.NoError(t, err)

	esClient, err := store.NewClient(
		log.NewNoop(), store.Config{}, store.WithClient(cli),
	)
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

	t.Run("should return error if the given urn is empty", func(t *testing.T) {
		params := asset.SoftDeleteAssetParams{
			URN:         "",
			UpdatedAt:   currentTime,
			RefreshedAt: currentTime,
			NewVersion:  asset.BaseVersion,
		}
		err = repo.SoftDeleteByURN(ctx, params)
		assert.ErrorIs(t, err, asset.ErrEmptyURN)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		ast := asset.Asset{
			ID:      "delete-id",
			URN:     "some-urn",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			Version: asset.BaseVersion,
		}

		err = repo.Upsert(ctx, ast)
		time.Sleep(1 * time.Second)
		require.NoError(t, err)

		// Ensure the asset exists before soft delete
		hits, total := searchByURN(t, cli, ast.URN)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, false, hits[0].IsDeleted)
		assert.Equal(t, asset.BaseVersion, hits[0].Version)

		params := asset.SoftDeleteAssetParams{
			URN:         ast.URN,
			UpdatedAt:   currentTime,
			RefreshedAt: currentTime,
			NewVersion:  "0.2",
		}
		err = repo.SoftDeleteByURN(ctx, params)
		assert.NoError(t, err)

		// Soft delete does not remove the asset, but marks it as deleted
		hits, total = searchByURN(t, cli, ast.URN)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, true, hits[0].IsDeleted)
		assert.Equal(t, "0.2", hits[0].Version)
		assert.Equal(t, currentTime, hits[0].UpdatedAt)
		assert.Equal(t, currentTime, hits[0].RefreshedAt)
	})

	t.Run("should ignore unavailable indices", func(t *testing.T) {
		ast1 := asset.Asset{
			ID:      "id1",
			Type:    asset.Type("table"),
			Service: bigqueryService,
			URN:     "test-urn1",
			Version: "0.1",
		}
		ast2 := asset.Asset{
			ID:      "id2",
			Type:    asset.Type("topic"),
			Service: kafkaService,
			URN:     "test-urn2",
			Version: "0.1",
		}
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		err = repo.Upsert(ctx, ast1)
		require.NoError(t, err)

		err = repo.Upsert(ctx, ast2)
		require.NoError(t, err)

		_, err = cli.Indices.Close([]string{kafkaService})
		require.NoError(t, err)

		params := asset.SoftDeleteAssetParams{
			URN:         ast1.URN,
			UpdatedAt:   currentTime,
			RefreshedAt: currentTime,
			NewVersion:  "0.2",
		}
		err = repo.SoftDeleteByURN(ctx, params)
		assert.NoError(t, err)
	})
}

func TestDiscoveryRepository_DeleteByQueryExpr(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		kafkaService    = "kafka-test"
	)

	cli, err := esTestServer.NewClient()
	require.NoError(t, err)

	esClient, err := store.NewClient(
		log.NewNoop(), store.Config{}, store.WithClient(cli),
	)
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

	t.Run("should return error if the given query expr is empty", func(t *testing.T) {
		queryExpr := asset.DeleteAssetExpr{
			ExprStr: queryexpr.ESExpr(""),
		}
		err = repo.DeleteByQueryExpr(ctx, queryExpr)
		assert.ErrorIs(t, err, asset.ErrEmptyQuery)
	})

	t.Run("should not return error on success", func(t *testing.T) {
		currentTime := time.Now().UTC()
		ast := asset.Asset{
			ID:          "delete-id",
			Type:        asset.Type("table"),
			Service:     bigqueryService,
			URN:         "some-urn",
			RefreshedAt: &currentTime,
		}

		err = repo.Upsert(ctx, ast)
		require.NoError(t, err)

		query := "refreshed_at <= '" + time.Now().Format("2006-01-02T15:04:05Z") +
			"' && service == '" + bigqueryService +
			"' && type == '" + asset.Type("table").String() + "'"
		queryExpr := asset.DeleteAssetExpr{
			ExprStr: queryexpr.ESExpr(query),
		}
		err = repo.DeleteByQueryExpr(ctx, queryExpr)
		assert.NoError(t, err)

		esQuery, _ := queryexpr.ValidateAndGetQueryFromExpr(queryExpr)

		res, err := cli.Search(
			cli.Search.WithBody(strings.NewReader(esQuery)),
			cli.Search.WithIndex("_all"),
		)
		require.NoError(t, err)
		assert.False(t, res.IsError())

		var body struct {
			Hits struct {
				Total elastic.TotalHits `json:"total"`
			} `json:"hits"`
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
		assert.Equal(t, int64(0), body.Hits.Total.Value)
	})

	t.Run("should ignore unavailable indices", func(t *testing.T) {
		currentTime := time.Now()
		ast1 := asset.Asset{
			ID:          "id1",
			Type:        asset.Type("table"),
			Service:     bigqueryService,
			URN:         "urn1",
			RefreshedAt: &currentTime,
		}
		ast2 := asset.Asset{
			ID:          "id2",
			Type:        asset.Type("topic"),
			Service:     kafkaService,
			URN:         "urn2",
			RefreshedAt: &currentTime,
		}
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		err = repo.Upsert(ctx, ast1)
		require.NoError(t, err)

		err = repo.Upsert(ctx, ast2)
		require.NoError(t, err)

		_, err = cli.Indices.Close([]string{kafkaService})
		require.NoError(t, err)

		query := "refreshed_at <= '" + time.Now().Format("2006-01-02T15:04:05Z") +
			"' && service == '" + kafkaService +
			"' && type == '" + asset.Type("topic").String() + "'"
		queryExpr := asset.DeleteAssetExpr{
			ExprStr: queryexpr.ESExpr(query),
		}
		err = repo.DeleteByQueryExpr(ctx, queryExpr)
		assert.NoError(t, err)
	})
}

func TestDiscoveryRepository_SoftDeleteAssets(t *testing.T) {
	var (
		ctx             = context.Background()
		bigqueryService = "bigquery-test"
		kafkaService    = "kafka-test"
		currentTime     = time.Now().UTC()
		userID          = "test-user-id"
	)

	cli, err := esTestServer.NewClient()
	require.NoError(t, err)

	esClient, err := store.NewClient(
		log.NewNoop(), store.Config{}, store.WithClient(cli),
	)
	require.NoError(t, err)

	repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

	t.Run("should not return error if the given assets is empty", func(t *testing.T) {
		err = repo.SoftDeleteAssets(ctx, []asset.Asset{}, false)
		assert.NoError(t, err)
	})

	t.Run("should not return error on success with do update version", func(t *testing.T) {
		ast := asset.Asset{
			ID:          "delete-id-1",
			Type:        asset.Type("table"),
			Service:     bigqueryService,
			URN:         "some-urn-1",
			Version:     "0.1",
			UpdatedAt:   currentTime,
			RefreshedAt: &currentTime,
			UpdatedBy:   user.User{ID: userID},
		}

		err = repo.Upsert(ctx, ast)
		time.Sleep(1 * time.Second)
		require.NoError(t, err)

		// Ensure the asset exists before soft delete
		hits, total := searchByURN(t, cli, ast.URN)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, false, hits[0].IsDeleted)
		assert.Equal(t, asset.BaseVersion, hits[0].Version)

		// Soft delete the asset
		err = repo.SoftDeleteAssets(ctx, []asset.Asset{ast}, true)
		time.Sleep(1 * time.Second)
		assert.NoError(t, err)

		// Soft delete does not remove the asset, but marks it as deleted
		hits, total = searchByURN(t, cli, ast.URN)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, true, hits[0].IsDeleted)
		assert.Equal(t, "0.2", hits[0].Version) // updated
		assert.Equal(t, currentTime, hits[0].UpdatedAt)
		assert.Equal(t, currentTime, hits[0].RefreshedAt)
	})

	t.Run("should not return error on success without do update version", func(t *testing.T) {
		ast := asset.Asset{
			ID:          "delete-id-2",
			Type:        asset.Type("table"),
			Service:     bigqueryService,
			URN:         "some-urn-2",
			Version:     "0.1",
			UpdatedAt:   currentTime,
			RefreshedAt: &currentTime,
			UpdatedBy:   user.User{ID: userID},
		}

		err = repo.Upsert(ctx, ast)
		time.Sleep(1 * time.Second)
		require.NoError(t, err)

		// Ensure the asset exists before soft delete
		hits, total := searchByURN(t, cli, ast.URN)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, false, hits[0].IsDeleted)
		assert.Equal(t, asset.BaseVersion, hits[0].Version)

		// Soft delete the asset
		err = repo.SoftDeleteAssets(ctx, []asset.Asset{ast}, false)
		time.Sleep(1000 * time.Millisecond)
		assert.NoError(t, err)

		// Soft delete does not remove the asset, but marks it as deleted
		hits, total = searchByURN(t, cli, ast.URN)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, true, hits[0].IsDeleted)
		assert.Equal(t, "0.1", hits[0].Version) // still the same version
		assert.Equal(t, currentTime, hits[0].UpdatedAt)
		assert.Equal(t, currentTime, hits[0].RefreshedAt)
	})

	t.Run("should ignore unavailable indices", func(t *testing.T) {
		ast1 := asset.Asset{
			ID:          "id1",
			Type:        asset.Type("table"),
			Service:     bigqueryService,
			URN:         "urn1",
			UpdatedAt:   currentTime,
			RefreshedAt: &currentTime,
			UpdatedBy:   user.User{ID: userID},
		}
		ast2 := asset.Asset{
			ID:          "id2",
			Type:        asset.Type("topic"),
			Service:     kafkaService,
			URN:         "urn2",
			UpdatedAt:   currentTime,
			RefreshedAt: &currentTime,
			UpdatedBy:   user.User{ID: userID},
		}
		cli, err := esTestServer.NewClient()
		require.NoError(t, err)
		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		err = repo.Upsert(ctx, ast1)
		require.NoError(t, err)

		err = repo.Upsert(ctx, ast2)
		require.NoError(t, err)

		_, err = cli.Indices.Close([]string{kafkaService})
		require.NoError(t, err)

		// Soft delete the asset
		err = repo.SoftDeleteAssets(ctx, []asset.Asset{ast1, ast2}, false)
		assert.NoError(t, err)
	})
}

func TestDiscoveryRepository_SyncAssets(t *testing.T) {
	t.Run("should return success", func(t *testing.T) {
		var (
			ctx       = context.Background()
			indexName = "bigquery-test"
		)

		cli, err := esTestServer.NewClient()
		require.NoError(t, err)

		_, err = cli.Indices.Create(indexName)
		require.NoError(t, err)

		esClient, err := store.NewClient(
			log.NewNoop(),
			store.Config{},
			store.WithClient(cli),
		)
		require.NoError(t, err)

		repo := store.NewDiscoveryRepository(esClient, log.NewNoop(), time.Second*10, []string{"number", "id"})

		cleanupFn, err := repo.SyncAssets(ctx, indexName)
		require.NoError(t, err)

		alias := cli.Indices.GetAlias
		resp, _ := alias(alias.WithIndex(indexName))
		require.NotEmpty(t, resp)

		err = cleanupFn()
		require.NoError(t, err)

		res, err := cli.Indices.Exists([]string{"bigquery-test"})
		require.Equal(t, res.StatusCode, 200)
		require.NoError(t, err)

		res, err = cli.Indices.Exists([]string{"bigquery-test-bak"})
		require.Equal(t, res.StatusCode, 404)
		require.NoError(t, err)
	})
}

func searchByURN(t *testing.T, cli *elasticsearch.Client, urn string) ([]searchByUrnFields, int64) {
	t.Helper()

	res, err := cli.Search(
		cli.Search.WithBody(strings.NewReader(fmt.Sprintf(`{"query":{"term":{"urn.keyword": %q}}}`, urn))),
		cli.Search.WithIndex("_all"),
	)
	require.NoError(t, err)
	assert.False(t, res.IsError())

	var result struct {
		Hits struct {
			Hits []struct {
				searchByUrnFields `json:"_source"`
			} `json:"hits"`
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}

	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))

	// Extract just the _source fields for easier assertion
	sources := make([]searchByUrnFields, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		sources[i] = hit.searchByUrnFields
	}

	return sources, result.Hits.Total.Value
}
