package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/testutils"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/goto/salt/log"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AssetRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	repository *postgres.AssetRepository
	userRepo   *postgres.UserRepository
	users      []user.User
	assets     []asset.Asset
	builder    sq.SelectBuilder
}

func (r *AssetRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus()
	r.client, err = newTestClient(r.T(), logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.userRepo, err = postgres.NewUserRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}

	r.repository, err = postgres.NewAssetRepository(r.client, r.userRepo, defaultGetMaxSize, defaultProviderName)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *AssetRepositoryTestSuite) createUsers(userRepo user.Repository) []user.User {
	var err error
	var users []user.User

	user1 := user.User{Email: "user-test-1@gotocompany.com", Provider: defaultProviderName}
	user1.ID, err = userRepo.Create(r.ctx, &user1)
	r.Require().NoError(err)
	users = append(users, user1)

	user2 := user.User{Email: "user-test-2@gotocompany.com", Provider: defaultProviderName}
	user2.ID, err = userRepo.Create(r.ctx, &user2)
	r.Require().NoError(err)
	users = append(users, user2)

	user3 := user.User{Email: "user-test-3@gotocompany.com", Provider: defaultProviderName}
	user3.ID, err = userRepo.Create(r.ctx, &user3)
	r.Require().NoError(err)
	users = append(users, user3)

	user4 := user.User{Email: "user-test-4@gotocompany.com", Provider: defaultProviderName}
	user4.ID, err = userRepo.Create(r.ctx, &user4)
	r.Require().NoError(err)
	users = append(users, user4)

	return users
}

func (r *AssetRepositoryTestSuite) BeforeTest(_, _ string) {
	err := testutils.RunMigrationsWithClient(r.T(), r.client)
	r.NoError(err)

	r.users = r.createUsers(r.userRepo)
	r.assets = r.insertRecord()
}

func (r *AssetRepositoryTestSuite) insertRecord() (assets []asset.Asset) {
	filePath := "./testdata/mock-asset-data.json"
	testFixtureJSON, err := os.ReadFile(filePath)
	if err != nil {
		return []asset.Asset{}
	}

	var data []asset.Asset
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return []asset.Asset{}
	}

	for _, d := range data {
		ast := asset.Asset{
			URN:         d.URN,
			Name:        d.Name,
			Type:        d.Type,
			Service:     d.Service,
			Description: d.Description,
			Data:        d.Data,
			Version:     asset.BaseVersion,
			UpdatedBy:   r.users[0],
		}

		insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(insertedAsset.ID)
		assets = append(assets, *insertedAsset)
	}

	return assets
}

func (r *AssetRepositoryTestSuite) TestBuildFilterQuery() {
	r.builder = sq.Select(`a.test as test`)

	testCases := []struct {
		description   string
		config        asset.Filter
		expectedQuery string
	}{
		{
			description: "should return sql query with types filter",
			config: asset.Filter{
				Types: []asset.Type{asset.Type("table")},
			},
			expectedQuery: `type IN ($1)`,
		},
		{
			description: "should return sql query with services filter",
			config: asset.Filter{
				Services: []string{"mysql", "kafka"},
			},
			expectedQuery: `service IN ($1,$2)`,
		},
		{
			description: "should return sql query with query fields filter",
			config: asset.Filter{
				QueryFields: []string{"name", "description"},
				Query:       "demo",
			},
			expectedQuery: `(name ILIKE $1 OR description ILIKE $2)`,
		},
		{
			description: "should return sql query with nested data query filter",
			config: asset.Filter{
				QueryFields: []string{"data.landscape.properties.project-id", "description"},
				Query:       "compass_002",
			},
			expectedQuery: `(data->'landscape'->'properties'->>'project-id' ILIKE $1 OR description ILIKE $2)`,
		},
		{
			// NOTE: Cannot have more than one key in map because golang's map does not guarantee order thus producing
			// inconsistent test.
			description: "should return sql query with asset's data fields filter",
			config: asset.Filter{
				Data: map[string][]string{
					"entity": {"gotocompany"},
				},
			},
			expectedQuery: `(data->>'entity' = $1)`,
		},
		{
			description: "should return sql query with asset's nested data fields filter",
			config: asset.Filter{
				Data: map[string][]string{
					"landscape.properties.project-id": {"compass_001"},
				},
			},
			expectedQuery: `(data->'landscape'->'properties'->>'project-id' = $1)`,
		},
		{
			description: "should return sql query with asset's nested multiple data fields filter ",
			config: asset.Filter{
				Data: map[string][]string{
					"properties.attributes.entity": {"alpha", "beta"},
				},
			},
			expectedQuery: `(data->'properties'->'attributes'->>'entity' = $1 OR data->'properties'->'attributes'->>'entity' = $2)`,
		},
		{
			description: "should return sql query with array index in query fields filter",
			config: asset.Filter{
				QueryFields: []string{"data.attributes.schema[0].name"},
				Query:       "column_alpha",
			},
			expectedQuery: `(data->'attributes'->'schema'->0->>'name' ILIKE $1)`,
		},
		{
			description: "should return sql query with array index in data filter",
			config: asset.Filter{
				Data: map[string][]string{
					"attributes.schema[0].name": {"column_alpha"},
				},
			},
			expectedQuery: `(data->'attributes'->'schema'->0->>'name' = $1)`,
		},
		{
			description: "should return sql query with array index accessing last element directly",
			config: asset.Filter{
				Data: map[string][]string{
					"attributes.schema[1].type": {"integer"},
				},
			},
			expectedQuery: `(data->'attributes'->'schema'->1->>'type' = $1)`,
		},
		{
			description: "should return sql query with array index as the terminal segment",
			config: asset.Filter{
				QueryFields: []string{"data.properties.dependencies[0]"},
				Query:       "test",
			},
			expectedQuery: `(data->'properties'->'dependencies'->>0 ILIKE $1)`,
		},
		{
			description: "should return sql query with dot-notation array index in query fields filter (API-safe format)",
			config: asset.Filter{
				QueryFields: []string{"data.attributes.schema.0.name"},
				Query:       "column_alpha",
			},
			expectedQuery: `(data->'attributes'->'schema'->0->>'name' ILIKE $1)`,
		},
		{
			description: "should return sql query with dot-notation array index in data filter (API-safe format)",
			config: asset.Filter{
				Data: map[string][]string{
					"attributes.schema.0.name": {"column_alpha"},
				},
			},
			expectedQuery: `(data->'attributes'->'schema'->0->>'name' = $1)`,
		},
		{
			description: "should return sql query with dot-notation array index as terminal segment",
			config: asset.Filter{
				QueryFields: []string{"data.properties.dependencies.0"},
				Query:       "test",
			},
			expectedQuery: `(data->'properties'->'dependencies'->>0 ILIKE $1)`,
		},
		{
			description: `should treat quoted integer segment as string key in query fields ("0" not an array index)`,
			config: asset.Filter{
				QueryFields: []string{`data.attributes."0".name`},
				Query:       "value",
			},
			expectedQuery: `(data->'attributes'->'0'->>'name' ILIKE $1)`,
		},
		{
			description: `should treat quoted integer segment as string key in data filter ("0" not an array index)`,
			config: asset.Filter{
				Data: map[string][]string{
					`attributes."0".name`: {"value"},
				},
			},
			expectedQuery: `(data->'attributes'->'0'->>'name' = $1)`,
		},
		{
			description: `should treat quoted integer as terminal string key`,
			config: asset.Filter{
				Data: map[string][]string{
					`attributes."0"`: {"value"},
				},
			},
			expectedQuery: `(data->'attributes'->>'0' = $1)`,
		},
	}

	for _, testCase := range testCases {
		r.Run(testCase.description, func() {
			result := r.repository.BuildFilterQuery(r.builder, testCase.config)
			query, _, err := result.ToSql()
			r.Require().NoError(err)
			query, err = sq.Dollar.ReplacePlaceholders(query)
			r.Require().NoError(err)

			actualQuery := strings.Split(query, "WHERE ")
			r.Equal(testCase.expectedQuery, actualQuery[1])
		})
	}
}

func (r *AssetRepositoryTestSuite) TestGetAll() {
	r.BeforeTest("", "")

	r.Run("should return error if SortBy key is invalid", func() {
		_, err := r.repository.GetAll(r.ctx, asset.Filter{
			SortBy: "wrong key name",
		})
		r.Require().NotNil(err)
		r.Require().ErrorContains(err, "error getting asset list")
	})

	r.Run("should return all assets without filtering based on size", func() {
		expectedSize := 15

		results, err := r.repository.GetAll(r.ctx, asset.Filter{})
		r.Require().NoError(err)
		r.Require().Len(results, expectedSize)
		for i := 0; i < expectedSize; i++ {
			r.assertAsset(&r.assets[i], &results[i])
		}
	})

	r.Run("should override default size using GetConfig.Size", func() {
		size := 6

		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Size: size,
		})
		r.Require().NoError(err)
		r.Require().Len(results, size)
		for i := 0; i < size; i++ {
			r.assertAsset(&r.assets[i], &results[i])
		}
	})

	r.Run("should fetch assets by offset defined in GetConfig.Offset", func() {
		offset := 2

		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Offset: offset,
		})
		r.Require().NoError(err)
		for i := offset; i > len(results)+offset; i++ {
			r.assertAsset(&r.assets[i], &results[i-offset])
		}
	})

	r.Run("should filter using type", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Types:         []asset.Type{"table"},
			SortBy:        "urn",
			SortDirection: "desc",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"twelfth-mock", "i-undefined-dfgdgd-avi", "e-test-grant2"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using service", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Services: []string{"mysql", "kafka"},
			SortBy:   "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"c-demo-kafka", "f-john-test-001", "i-test-grant", "i-undefined-dfgdgd-avi"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using query fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			QueryFields: []string{"name", "description"},
			Query:       "demo",
			SortBy:      "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"c-demo-kafka", "e-test-grant2"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter only using nested query data fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			QueryFields: []string{"data.landscape.properties.project-id", "data.title"},
			Query:       "compass_001",
			SortBy:      "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"i-test-grant", "j-xcvcx"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using query field with nested query data fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			QueryFields: []string{"data.landscape.properties.project-id", "description"},
			Query:       "compass_002",
			SortBy:      "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"g-jane-kafka-1a", "h-test-new-kafka"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using asset's data fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"entity":  {"gotocompany"},
				"country": {"th"},
			},
		})
		r.Require().NoError(err)

		expectedURNs := []string{"e-test-grant2", "h-test-new-kafka", "i-test-grant"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using asset's nested data fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"landscape.properties.project-id": {"compass_001"},
				"country":                         {"vn"},
			},
		})
		r.Require().NoError(err)

		expectedURNs := []string{"j-xcvcx"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using asset's nonempty data fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"properties.dependencies": {"_nonempty"},
			},
		})
		r.Require().NoError(err)

		expectedURNs := []string{"nine-mock", "eleven-mock"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using asset's different nonempty data fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"properties.dependencies": {"_nonempty"},
				"entity":                  {"gotocompany"},
				"urn":                     {"j-xcvcx"},
				"country":                 {"vn"},
			},
		})
		r.Require().NoError(err)

		expectedURNs := []string{"nine-mock"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using array index in data fields matching both records", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"attributes.schema[0].name": {"column_alpha"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-1", "array-index-mock-2"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using array index in data fields matching only first record", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"attributes.schema[1].name": {"column_beta"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-1"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using array index in data fields matching only second record", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"attributes.schema[1].name": {"column_gamma"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-2"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using array index in query fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			QueryFields: []string{"data.attributes.schema[0].name"},
			Query:       "column_alpha",
			SortBy:      "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-1", "array-index-mock-2"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using array index in data fields combined with other data filters", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"attributes.schema[0].name": {"column_alpha"},
				"attributes.schema[1].type": {"integer"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-1"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	// dot-notation (attributes.schema.0.name) is the API-safe equivalent of bracket-notation
	// (attributes.schema[0].name). When passing through HTTP query params, grpc-gateway's
	// bracket parser misinterprets nested brackets (e.g. data[schema[0].name]), so dot-notation
	// must be used instead: data[schema.0.name]=value.
	r.Run("should filter using dot-notation array index in data fields matching both records", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"attributes.schema.0.name": {"column_alpha"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-1", "array-index-mock-2"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using dot-notation array index in data fields matching only first record", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				"attributes.schema.1.name": {"column_beta"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-1"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run("should filter using dot-notation array index in query fields", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			QueryFields: []string{"data.attributes.schema.0.name"},
			Query:       "column_alpha",
			SortBy:      "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"array-index-mock-1", "array-index-mock-2"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	// Quoted segment: wrap an integer-looking key in double-quotes to force string-key lookup.
	// e.g. attributes."0" targets the JSON object key "0", not array index 0.
	r.Run(`should filter using quoted integer segment as string key in data fields`, func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				`attributes."0"`: {"numeric-string-value"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)

		expectedURNs := []string{"quoted-key-mock"}
		r.Equal(len(expectedURNs), len(results))
		for i := range results {
			r.Equal(expectedURNs[i], results[i].URN)
		}
	})

	r.Run(`should not match array index when using quoted string key`, func() {
		// array-index-mock-1 has attributes.schema[0] (array), not attributes."0"
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Data: map[string][]string{
				`attributes."0"`: {"column_alpha"},
			},
			SortBy: "urn",
		})
		r.Require().NoError(err)
		r.Equal(0, len(results))
	})
}

func (r *AssetRepositoryTestSuite) TestGetTypes() {
	r.BeforeTest("", "")

	type testCase struct {
		Description string
		Filter      asset.Filter
		Expected    map[asset.Type]int
	}

	const (
		typeDashboard = asset.Type("dashboard")
		typeJob       = asset.Type("job")
		typeTable     = asset.Type("table")
		typeTopic     = asset.Type("topic")
	)

	testCases := []testCase{
		{
			Description: "should return maps of asset count without filter",
			Filter:      asset.Filter{},
			Expected: map[asset.Type]int{
				typeDashboard: 5,
				typeJob:       1,
				typeTable:     3,
				typeTopic:     6,
			},
		},
		{
			Description: "should filter using service",
			Filter: asset.Filter{
				Services: []string{"mysql", "kafka"},
				SortBy:   "urn",
			},
			Expected: map[asset.Type]int{
				typeTable: 1,
				typeTopic: 3,
			},
		},
		{
			Description: "should filter using query fields",
			Filter: asset.Filter{
				QueryFields: []string{"name", "description"},
				Query:       "demo",
				SortBy:      "urn",
			},
			Expected: map[asset.Type]int{
				typeTable: 1,
				typeTopic: 1,
			},
		},
		{
			Description: "should filter only using nested query data fields",
			Filter: asset.Filter{
				QueryFields: []string{"data.landscape.properties.project-id", "data.title"},
				Query:       "compass_001",
				SortBy:      "urn",
			},
			Expected: map[asset.Type]int{
				typeDashboard: 1,
				typeTopic:     1,
			},
		},

		{
			Description: "should filter using query field with nested query data fields",
			Filter: asset.Filter{
				QueryFields: []string{"data.landscape.properties.project-id", "description"},
				Query:       "compass_002",
				SortBy:      "urn",
			},
			Expected: map[asset.Type]int{
				typeDashboard: 1,
				typeJob:       1,
			},
		},
		{
			Description: "should filter using asset's data fields",
			Filter: asset.Filter{
				Data: map[string][]string{
					"entity":  {"gotocompany"},
					"country": {"th"},
				},
			},
			Expected: map[asset.Type]int{
				typeJob:   1,
				typeTable: 1,
				typeTopic: 1,
			},
		},
		{
			Description: "should filter using asset's nested data fields",
			Filter: asset.Filter{
				Data: map[string][]string{
					"landscape.properties.project-id": {"compass_001"},
					"country":                         {"vn"},
				},
			},
			Expected: map[asset.Type]int{
				typeDashboard: 1,
			},
		},
		{
			Description: "should filter using asset's nonempty data fields",
			Filter: asset.Filter{
				Data: map[string][]string{
					"properties.dependencies": {"_nonempty"},
				},
			},
			Expected: map[asset.Type]int{
				typeDashboard: 2,
			},
		},
		{
			Description: "should filter using asset's different nonempty data fields",
			Filter: asset.Filter{
				Data: map[string][]string{
					"properties.dependencies": {"_nonempty"},
					"entity":                  {"gotocompany"},
					"urn":                     {"j-xcvcx"},
					"country":                 {"vn"},
				},
			},
			Expected: map[asset.Type]int{
				typeDashboard: 1,
			},
		},
	}

	for _, tc := range testCases {
		r.Run(tc.Description, func() {
			typeMap, err := r.repository.GetTypes(context.Background(), tc.Filter)
			r.NoError(err)

			keys := make([]string, 0, len(typeMap))

			for k := range typeMap {
				keys = append(keys, k.String())
			}
			sort.Strings(keys)

			sortedMap := make(map[asset.Type]int)
			for _, k := range keys {
				sortedMap[asset.Type(k)] = typeMap[asset.Type(k)]
			}

			if !cmp.Equal(tc.Expected, sortedMap) {
				r.T().Fatalf("expected is %+v but got %+v", tc.Expected, sortedMap)
			}
		})
	}
}

func (r *AssetRepositoryTestSuite) TestGetCount() {
	// populate assets
	total := 12
	typ := asset.Type("job")
	service := []string{"service-getcount"}
	for i := 0; i < total; i++ {
		ast := asset.Asset{
			URN:       fmt.Sprintf("urn-getcount-%d", i),
			Name:      fmt.Sprintf("getcount-%d", i),
			Type:      typ,
			Service:   service[0],
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(insertedAsset.ID)
	}

	r.Run("should return total assets with filter", func() {
		actual, err := r.repository.GetCount(r.ctx, asset.Filter{
			Types:    []asset.Type{typ},
			Services: service,
		})
		r.Require().NoError(err)
		r.Equal(total, actual)
	})
}

func (r *AssetRepositoryTestSuite) TestGetCountByQueryExpr() {
	ast := asset.Asset{
		URN:       uuid.NewString() + "urn-gcqe-1",
		Name:      "gcqe-1",
		Type:      "table",
		Service:   "bigquery",
		UpdatedBy: r.users[0],
		Data:      map[string]interface{}{},
	}
	_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
	r.Require().NoError(err)

	r.Run("should return count matching query expression", func() {
		count, err := r.repository.GetCountByQueryExpr(r.ctx, asset.DeleteAssetExpr{ExprStr: queryexpr.SQLExpr(fmt.Sprintf("urn == '%s' && refreshed_at != nil && type == '%s' && service == '%s'\n", ast.URN, "table", "bigquery"))})
		r.Require().NoError(err)
		r.Equal(uint32(1), count)
	})

	r.Run("should return error on invalid query expression", func() {
		_, err := r.repository.GetCountByQueryExpr(r.ctx, asset.DeleteAssetExpr{ExprStr: queryexpr.SQLExpr("invalid!!!")})
		r.Error(err)
	})
}

func (r *AssetRepositoryTestSuite) TestGetCountByIsDeletedAndServicesAndUpdatedAt() {
	now := time.Now()
	serviceA := "serviceA"
	serviceB := "serviceB"

	asset1 := asset.Asset{
		URN:       "urn:svcA:1",
		Name:      "Asset 1",
		Service:   serviceA,
		Type:      "topic",
		IsDeleted: false,
		UpdatedBy: r.users[0],
		Data:      map[string]interface{}{},
	}
	asset2 := asset.Asset{
		URN:       "urn:svcA:2",
		Name:      "Asset 2",
		Service:   serviceA,
		Type:      "topic",
		IsDeleted: true,
		UpdatedBy: r.users[0],
		Data:      map[string]interface{}{},
	}
	asset3 := asset.Asset{
		URN:       "urn:svcB:1",
		Name:      "Asset 3",
		Service:   serviceB,
		Type:      "topic",
		IsDeleted: false,
		UpdatedBy: r.users[0],
		Data:      map[string]interface{}{},
	}

	_, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
	r.Require().NoError(err)
	_, _, err = r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
	r.Require().NoError(err)
	_, _, err = r.repository.Upsert(r.ctx, &asset3, false, asset.Config{})
	r.Require().NoError(err)

	thresholdUpdatedAt := now.Add(time.Hour)

	r.Run("should return correct count for isDeleted=false and serviceA", func() {
		count, err := r.repository.GetCountByIsDeletedAndServicesAndUpdatedAt(r.ctx, false, []string{serviceA}, thresholdUpdatedAt)
		r.NoError(err)
		r.Equal(uint32(1), count)
	})

	r.Run("should return correct count for isDeleted=true and serviceA", func() {
		count, err := r.repository.GetCountByIsDeletedAndServicesAndUpdatedAt(r.ctx, true, []string{serviceA}, thresholdUpdatedAt)
		r.NoError(err)
		r.Equal(uint32(1), count)
	})

	r.Run("should return correct count for isDeleted=false and serviceB", func() {
		count, err := r.repository.GetCountByIsDeletedAndServicesAndUpdatedAt(r.ctx, false, []string{serviceB}, thresholdUpdatedAt)
		r.NoError(err)
		r.Equal(uint32(1), count)
	})

	r.Run("should return 0 for isDeleted=true and serviceB", func() {
		count, err := r.repository.GetCountByIsDeletedAndServicesAndUpdatedAt(r.ctx, true, []string{serviceB}, thresholdUpdatedAt)
		r.NoError(err)
		r.Equal(uint32(0), count)
	})

	r.Run("should return error for empty services", func() {
		_, err := r.repository.GetCountByIsDeletedAndServicesAndUpdatedAt(r.ctx, false, []string{}, thresholdUpdatedAt)
		r.ErrorIs(err, asset.ErrEmptyServices)
	})

	r.Run("should return correct count for all services config", func() {
		allServices := []string{asset.AllServicesCleanupConfig}
		count, err := r.repository.GetCountByIsDeletedAndServicesAndUpdatedAt(r.ctx, false, allServices, thresholdUpdatedAt)
		r.NoError(err)
		r.GreaterOrEqual(count, uint32(2))
	})

	// cleanup
	err = r.repository.DeleteByURN(r.ctx, asset1.URN)
	r.NoError(err)
	err = r.repository.DeleteByURN(r.ctx, asset2.URN)
	r.NoError(err)
	err = r.repository.DeleteByURN(r.ctx, asset3.URN)
	r.NoError(err)
}

func (r *AssetRepositoryTestSuite) TestGetByID() {
	r.Run("return error from client if asset not an uuid", func() {
		_, err := r.repository.GetByID(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
	})

	r.Run("return NotFoundError if asset does not exist", func() {
		uuidTest := "2aabb450-f986-44e2-a6db-7996861d5004"
		_, err := r.repository.GetByID(r.ctx, uuidTest)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuidTest})
	})

	r.Run("return correct asset from db", func() {
		asset1 := asset.Asset{
			URN:       "urn-gbi-1",
			Name:      "gbi-1",
			Type:      "table",
			Service:   "bigquery",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:       "urn-gbi-2",
			Name:      "gbi-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}

		var err error
		insertedAsset, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)
		r.NotEmpty(insertedAsset.ID)

		insertedAsset2, _, err := r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)
		r.NotEmpty(insertedAsset2.ID)

		result, err := r.repository.GetByID(r.ctx, insertedAsset2.ID)
		r.NoError(err)
		asset2.UpdatedBy = r.users[1]
		r.assertAsset(insertedAsset2, &result)
	})

	r.Run("return owners if any", func() {
		ast := asset.Asset{
			URN:     "urn-gbi-3",
			Name:    "gbi-3",
			Type:    "table",
			Service: "bigquery",
			Owners: []user.User{
				r.users[1],
				r.users[2],
			},
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}

		insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(insertedAsset.ID)

		result, err := r.repository.GetByID(r.ctx, insertedAsset.ID)
		r.NoError(err)
		r.Len(result.Owners, len(ast.Owners))
		for i, owner := range result.Owners {
			r.Equal(ast.Owners[i].ID, owner.ID)
		}
	})
}

func (r *AssetRepositoryTestSuite) TestGetByURN() {
	r.Run("return NotFoundError if asset does not exist", func() {
		urn := "urn-gbi-1"
		_, err := r.repository.GetByURN(r.ctx, urn)
		r.ErrorAs(err, &asset.NotFoundError{URN: urn})
	})

	r.Run("return correct asset from db", func() {
		asset1 := asset.Asset{
			URN:       "urn-gbi-1",
			Name:      "gbi-1",
			Type:      "table",
			Service:   "bigquery",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:       "urn-gbi-2",
			Name:      "gbi-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}

		insertedAsset, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)
		r.NotEmpty(insertedAsset.ID)

		insertedAsset2, _, err := r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)
		r.NotEmpty(insertedAsset2.ID)

		result, err := r.repository.GetByURN(r.ctx, "urn-gbi-2")
		r.NoError(err)
		r.assertAsset(insertedAsset2, &result)
	})

	r.Run("return owners if any", func() {
		ast := asset.Asset{
			URN:     "urn-gbi-3",
			Name:    "gbi-3",
			Type:    "table",
			Service: "bigquery",
			Owners: []user.User{
				r.users[1],
				r.users[2],
			},
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}

		_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)

		result, err := r.repository.GetByURN(r.ctx, ast.URN)
		r.NoError(err)
		r.Len(result.Owners, len(ast.Owners))
		for i, owner := range result.Owners {
			r.Equal(ast.Owners[i].ID, owner.ID)
		}
	})
}

func (r *AssetRepositoryTestSuite) TestVersions() {
	currentTime := time.Date(2024, time.August, 20, 8, 19, 49, 0, time.UTC)
	assetURN := uuid.NewString() + "urn-u-2-version"
	excludedChangelogPaths := []string{"data.update_time"}
	// v0.1
	astVersioning := asset.Asset{
		URN:         assetURN,
		Name:        "u-2-version",
		Type:        "table",
		Service:     "bigquery",
		UpdatedBy:   r.users[1],
		RefreshedAt: &currentTime,
		Data:        map[string]interface{}{},
	}

	insertedAsset, _, err := r.repository.Upsert(r.ctx, &astVersioning, false, asset.Config{})
	r.Require().NoError(err)
	r.Require().NotEmpty(insertedAsset.ID)
	astVersioning.ID = insertedAsset.ID

	// v0.2
	astVersioning.Description = "new description in v0.2"
	upsertedAsset, _, err := r.repository.Upsert(r.ctx, &astVersioning, false, asset.Config{})
	r.Require().NoError(err)
	r.Require().Equal(upsertedAsset.ID, astVersioning.ID)

	// v0.3
	astVersioning.Owners = []user.User{
		{
			Email: "user@gotocompany.com",
		},
		{
			Email:    "meteor@gotocompany.com",
			Provider: "meteor",
		},
	}
	upsertedAsset, _, err = r.repository.Upsert(r.ctx, &astVersioning, false, asset.Config{})
	r.Require().NoError(err)
	r.Require().Equal(upsertedAsset.ID, astVersioning.ID)

	// v0.4
	astVersioning.Data = map[string]interface{}{
		"data1": float64(12345),
	}
	upsertedAsset, _, err = r.repository.Upsert(r.ctx, &astVersioning, false, asset.Config{})
	r.Require().NoError(err)
	r.Require().Equal(upsertedAsset.ID, astVersioning.ID)

	// v0.5
	astVersioning.Labels = map[string]string{
		"key1": "value1",
	}

	upsertedAsset, _, err = r.repository.Upsert(r.ctx, &astVersioning, false, asset.Config{})
	r.Require().NoError(err)
	r.Require().Equal(upsertedAsset.ID, astVersioning.ID)

	r.Run("should return current version of an assets", func() {
		expected := asset.Asset{
			ID:          astVersioning.ID,
			URN:         assetURN,
			Name:        "u-2-version",
			Type:        "table",
			Service:     "bigquery",
			Description: "new description in v0.2",
			Data:        map[string]interface{}{"data1": float64(12345)},
			Labels:      map[string]string{"key1": "value1"},
			Version:     "0.5",
			UpdatedBy:   r.users[1],
			RefreshedAt: &currentTime,
		}

		ast, err := r.repository.GetByID(r.ctx, astVersioning.ID)
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners := ast.Owners
		ast.Owners = nil
		r.NoError(err)
		// making updatedby user time empty to make ast comparable
		clearTimestamps(&ast)
		r.Equal(expected, ast)

		r.Len(astOwners, 2)
	})

	r.Run("should return invalid if invalid uuid is set", func() {
		_, err := r.repository.GetByVersionWithID(r.ctx, "invalid uuid", "0.5")
		r.ErrorIs(err, asset.InvalidError{AssetID: "invalid uuid"})
	})

	r.Run("should return current version of an assets with by version", func() {
		expected := asset.Asset{
			ID:          astVersioning.ID,
			URN:         assetURN,
			Name:        "u-2-version",
			Type:        "table",
			Service:     "bigquery",
			Description: "new description in v0.2",
			Data:        map[string]interface{}{"data1": float64(12345)},
			Labels:      map[string]string{"key1": "value1"},
			Version:     "0.5",
			UpdatedBy:   r.users[1],
			RefreshedAt: &currentTime,
		}

		ast, err := r.repository.GetByVersionWithID(r.ctx, astVersioning.ID, "0.5")
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners := ast.Owners
		ast.Owners = nil
		r.NoError(err)
		// making updatedby user time empty to make ast comparable
		clearTimestamps(&ast)
		r.Equal(expected, ast)

		r.Len(astOwners, 2)

		ast, err = r.repository.GetByVersionWithURN(r.ctx, astVersioning.URN, "0.5")
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners = ast.Owners
		ast.Owners = nil
		r.NoError(err)
		// making updatedby user time empty to make ast comparable
		clearTimestamps(&ast)
		r.Equal(expected, ast)
		r.Len(astOwners, 2)
	})

	r.Run("should return a specific version of an asset", func() {
		version := "0.3"
		expected := asset.Asset{
			ID:          astVersioning.ID,
			URN:         assetURN,
			Name:        "u-2-version",
			Type:        "table",
			Service:     "bigquery",
			Description: "new description in v0.2",
			Version:     "0.3",
			Changelog: diff.Changelog{
				diff.Change{Type: "create", Path: []string{"owners", "0", "email"}, From: interface{}(nil), To: "user@gotocompany.com"},
				diff.Change{Type: "create", Path: []string{"owners", "1", "email"}, From: interface{}(nil), To: "meteor@gotocompany.com"},
			},
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}
		expectedOwners := []user.User{
			{
				Email: "user@gotocompany.com",
			},
			{
				Email:    "meteor@gotocompany.com",
				Provider: "meteor",
			},
		}
		astVer, err := r.repository.GetByVersionWithID(r.ctx, astVersioning.ID, version)
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners := astVer.Owners
		astVer.Owners = nil
		r.Assert().NoError(err)
		// making updatedby user time empty to make ast comparable
		clearTimestamps(&astVer)
		r.Assert().Equal(expected, astVer)

		for i := 0; i < len(astOwners); i++ {
			astOwners[i].ID = ""
		}
		r.Assert().Equal(expectedOwners, astOwners)

		astVer, err = r.repository.GetByVersionWithURN(r.ctx, astVersioning.URN, version)
		// hard to get the internally generated user id, we exclude the owners from the assertion
		astOwners = astVer.Owners
		astVer.Owners = nil
		r.Assert().NoError(err)
		// making updatedby user time empty to make ast comparable
		clearTimestamps(&astVer)
		r.Assert().Equal(expected, astVer)

		for i := 0; i < len(astOwners); i++ {
			astOwners[i].ID = ""
		}
		r.Assert().Equal(expectedOwners, astOwners)
	})

	r.Run("should return 3 last versions of an assets if there are exist", func() {
		assetURN := uuid.NewString() + "urn-u-3-version"
		// v0.1
		ast := asset.Asset{
			URN:       assetURN,
			Name:      "u-3-version",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}
		insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(insertedAsset.ID)
		ast.ID = insertedAsset.ID

		for i := 2; i < 100; i++ {
			ast.Description = "new description in v0." + strconv.Itoa(i)
			upsertedAsset, _, err = r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().Equal(upsertedAsset.ID, ast.ID)
		}

		expected := []asset.Asset{
			{
				ID:      ast.ID,
				URN:     assetURN,
				Type:    "table",
				Service: "bigquery",
				Version: "0.13",
				Changelog: diff.Changelog{
					diff.Change{Type: "update", Path: []string{"description"}, From: "new description in v0.12", To: "new description in v0.13"},
				},
				UpdatedBy: r.users[1],
			},
			{
				ID:      ast.ID,
				URN:     assetURN,
				Type:    "table",
				Service: "bigquery",
				Version: "0.12",
				Changelog: diff.Changelog{
					diff.Change{Type: "update", Path: []string{"description"}, From: "new description in v0.11", To: "new description in v0.12"},
				},
				UpdatedBy: r.users[1],
			},
			{
				ID:      ast.ID,
				URN:     assetURN,
				Type:    "table",
				Service: "bigquery",
				Version: "0.11",
				Changelog: diff.Changelog{
					diff.Change{Type: "update", Path: []string{"description"}, From: "new description in v0.10", To: "new description in v0.11"},
				},
				UpdatedBy: r.users[1],
			},
		}

		assetVersions, err := r.repository.GetVersionHistory(r.ctx, asset.Filter{Size: 3, Offset: 86}, ast.ID, excludedChangelogPaths)
		r.NoError(err)
		// making updatedby user time empty to make ast comparable
		for i := 0; i < len(assetVersions); i++ {
			clearTimestamps(&assetVersions[i])
		}
		r.Equal(expected, assetVersions)
	})

	r.Run("should return asset count equal to defaultGetMaxSize if Size is set to 0", func() {
		assetURN := uuid.NewString() + "urn-u-3-version"
		// v0.1
		ast := asset.Asset{
			URN:       assetURN,
			Name:      "u-3-version",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: r.users[1],
			Data:      map[string]interface{}{},
		}
		insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(insertedAsset.ID)
		ast.ID = insertedAsset.ID

		for i := 2; i < 100; i++ {
			ast.Description = "new description in v0." + strconv.Itoa(i)
			upsertedAsset, _, err = r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().Equal(upsertedAsset.ID, ast.ID)
		}

		assetVersions, err := r.repository.GetVersionHistory(r.ctx, asset.Filter{Size: 0, Offset: 86}, ast.ID, excludedChangelogPaths)
		r.NoError(err)
		r.Equal(defaultGetMaxSize, len(assetVersions))
	})

	r.Run("should return error if invalid uuid is passed", func() {
		assetURN := "invalid uuid"
		_, err := r.repository.GetVersionHistory(r.ctx, asset.Filter{Size: 3, Offset: 86}, assetURN, excludedChangelogPaths)
		r.NotNil(err)
		r.Equal(asset.InvalidError{AssetID: assetURN}, err)
	})

	r.Run("should return error from GetByVersionWithID when version does not exist", func() {
		_, err := r.repository.GetByVersionWithID(r.ctx, astVersioning.ID, "99.99")
		r.Error(err)
	})

	r.Run("should return error from GetByVersionWithURN when version does not exist", func() {
		_, err := r.repository.GetByVersionWithURN(r.ctx, astVersioning.URN, "99.99")
		r.Error(err)
	})

	r.Run("should return error from GetByVersionWithURN when URN does not exist", func() {
		_, err := r.repository.GetByVersionWithURN(r.ctx, "non-existent-urn", "0.1")
		r.Error(err)
	})
}

func (r *AssetRepositoryTestSuite) TestNewAssetRepository() {
	r.Run("should return error when client is nil", func() {
		repo, err := postgres.NewAssetRepository(nil, r.userRepo, 0, "shield")
		r.Error(err)
		r.Nil(repo)
	})

	r.Run("should use default max size when zero is passed", func() {
		repo, err := postgres.NewAssetRepository(r.client, r.userRepo, 0, "shield")
		r.Require().NoError(err)
		r.NotNil(repo)
	})

	r.Run("should use default provider when empty is passed", func() {
		repo, err := postgres.NewAssetRepository(r.client, r.userRepo, 10, "")
		r.Require().NoError(err)
		r.NotNil(repo)
	})
}

func (r *AssetRepositoryTestSuite) TestUpsert() {
	refreshedAtTime := time.Date(2024, time.August, 20, 8, 19, 49, 0, time.UTC)
	currentTime := time.Now().UTC()
	r.Run("on insert", func() {
		r.Run("should return error when URN is empty", func() {
			ast := asset.Asset{Name: "name", Type: "table", Service: "bigquery", Data: map[string]interface{}{}, UpdatedBy: r.users[0]}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.ErrorContains(err, "urn is required")
		})

		r.Run("should return error when type is empty", func() {
			ast := asset.Asset{URN: uuid.NewString(), Name: "name", Service: "bigquery", Data: map[string]interface{}{}, UpdatedBy: r.users[0]}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.ErrorContains(err, "type is required")
		})

		r.Run("should return error when type is invalid", func() {
			ast := asset.Asset{URN: uuid.NewString(), Name: "name", Type: "invalid-type", Service: "bigquery", Data: map[string]interface{}{}, UpdatedBy: r.users[0]}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.ErrorContains(err, "type is invalid")
		})

		r.Run("should return error when name is empty", func() {
			ast := asset.Asset{URN: uuid.NewString(), Type: "table", Service: "bigquery", Data: map[string]interface{}{}, UpdatedBy: r.users[0]}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.ErrorContains(err, "name is required")
		})

		r.Run("should return error when data is nil", func() {
			ast := asset.Asset{URN: uuid.NewString(), Name: "name", Type: "table", Service: "bigquery", UpdatedBy: r.users[0]}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.ErrorContains(err, "data is required")
		})

		r.Run("should return error when service is empty", func() {
			ast := asset.Asset{URN: uuid.NewString(), Name: "name", Type: "table", Data: map[string]interface{}{}, UpdatedBy: r.users[0]}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.ErrorContains(err, "service is required")
		})

		r.Run("set ID to asset and version to base version", func() {
			ast := asset.Asset{
				URN:         uuid.NewString() + "urn-u-1",
				Name:        "u-1",
				Type:        "table",
				Service:     "bigquery",
				URL:         "https://sample-url.com",
				UpdatedBy:   r.users[0],
				RefreshedAt: &refreshedAtTime,
				Data:        map[string]interface{}{},
			}
			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)
			r.Equal(asset.BaseVersion, insertedAsset.Version)
			r.NotEmpty(insertedAsset.ID)
			r.NotEmpty(ast.CreatedAt)
			r.NotEmpty(ast.UpdatedAt)
			r.NotEqual(time.Time{}, insertedAsset.CreatedAt)
			r.NotEqual(time.Time{}, insertedAsset.UpdatedAt)

			ast.ID = insertedAsset.ID
			ast.Version = asset.BaseVersion
			r.assertAsset(&ast, insertedAsset)

			ast2 := ast
			ast2.RefreshedAt = nil
			ast2.Description = "create a new version" // to force fetch from asset_versions.
			_, _, err = r.repository.Upsert(r.ctx, &ast2, false, asset.Config{})
			r.NoError(err)
			r.Greater(ast2.UpdatedAt.UnixNano(), ast.UpdatedAt.UnixNano())
			assetv1, err := r.repository.GetByVersionWithID(r.ctx, ast.ID, asset.BaseVersion)
			r.NoError(err)
			r.Equal("0.1", assetv1.Version)
		})

		r.Run("should store owners if any", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-3",
				Name:    "u-3",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					r.users[1],
					{Email: r.users[2].Email},
					{ID: r.users[1].ID}, // should get deduplicated by ID
				},
				UpdatedBy: r.users[0],
				Data:      map[string]interface{}{},
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotEmpty(insertedAsset.ID)
			r.Len(insertedAsset.Owners, 2)
			r.Equal(r.users[1].ID, insertedAsset.Owners[0].ID)
			r.Equal(r.users[2].ID, insertedAsset.Owners[1].ID)
		})

		r.Run("should create owners as users if they do not exist yet", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-3a",
				Name:    "u-3a",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					{Email: "newuser@example.com"},
					{Email: "newuser2@example.com"},
					{Email: "newuser@example.com"}, // should get deduplicated by ID on fetch user by email
				},
				UpdatedBy: r.users[0],
				Data:      map[string]interface{}{},
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			r.Len(insertedAsset.Owners, 2)
			r.Equal(ast.Owners[0].Email, insertedAsset.Owners[0].Email)
		})

		r.Run("should initialize sql_version when sql is present", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-init-sql",
				Name:      "optimus-init-sql",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)
			optimus := insertedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Nil(optimus["resolved_sql_version"])
		})

		r.Run("should initialize both versions when sql and resolved_sql are present", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-optimus-init-both",
				Name:    "optimus-init-both",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{"optimus": map[string]interface{}{
					"sql":          "SELECT 1",
					"resolved_sql": "SELECT 2",
				}},
				UpdatedBy: r.users[0],
			}
			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)
			optimus := insertedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})

		r.Run("should call column lineage producer on insert when host provided", func() {
			// setup mock server to assert request and return a valid lineage response
			called := false
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/api/v1/lineage/columns" {
					w.WriteHeader(404)
					return
				}
				var payload map[string]string
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					w.WriteHeader(400)
					return
				}
				// ensure query exists in payload
				if _, ok := payload["query"]; !ok {
					w.WriteHeader(400)
					return
				}
				called = true
				resp := map[string]interface{}{
					"target_table": "project.schema.table",
					"columns": []map[string]interface{}{
						{
							"target_column": "jakarta_transaction_date",
							"sources":       []map[string]string{{"table": "pproject.schema.upstream_table", "column": "date_detail"}},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer srv.Close()

			// first insert baseline asset with only sql so update flow will produce a changelog
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-optimus-produce-on-insert",
				Name:    "optimus-produce-on-insert",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{"optimus": map[string]interface{}{
					"sql": "SELECT 1",
				}},
				UpdatedBy: r.users[0],
			}

			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			// now update asset to include resolved_sql which should generate a changelog and producer
			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 1",
				"resolved_sql": "SELECT 1 resolved",
			}}

			cfg := asset.Config{ColumnLineageHost: srv.URL, ColumnLineageChangeIdentifier: "data.optimus.resolved_sql"}
			updatedAsset, producer, err := r.repository.Upsert(r.ctx, &ast, false, cfg)
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			r.Require().NotNil(producer)

			// call producer which should perform HTTP request to mock server
			g, err := producer(context.Background())
			r.Require().NoError(err)
			r.True(called, "expected lineage service to be called")
			r.NotNil(g)
			r.Greater(len(g), 0)
		})

		r.Run("should return error when lineage service returns 500 and context timeout", func() {
			// server that returns 500
			srvErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(500)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to produce column lineage"})
			}))
			defer srvErr.Close()

			// first insert baseline asset with only sql so update flow will produce a changelog
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-optimus-produce-err",
				Name:    "optimus-produce-err",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{"optimus": map[string]interface{}{
					"sql": "SELECT 1",
				}},
				UpdatedBy: r.users[0],
			}

			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			// now update asset to include resolved_sql which should generate a changelog and producer
			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 1",
				"resolved_sql": "SELECT 1 resolved",
			}}

			cfg := asset.Config{ColumnLineageHost: srvErr.URL, ColumnLineageChangeIdentifier: "data.optimus.resolved_sql"}
			updatedAsset, producer, err := r.repository.Upsert(r.ctx, &ast, false, cfg)
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			r.Require().NotNil(producer)

			// call producer which should perform HTTP request to mock server and return an error
			_, err = producer(context.Background())
			r.Require().Error(err)

			// server that blocks until signaled; using a blocking handler ensures the client
			// request will be waiting and a canceled context will deterministically cancel it.
			blockCh := make(chan struct{})
			srvSlow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				<-blockCh
				w.WriteHeader(200)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"target_table": "a.b.c", "columns": []interface{}{}})
			}))
			defer srvSlow.Close()

			// modify both sql and resolved_sql so changelog includes resolved_sql and sql_version
			// will be bumped, ensuring a ColumnLineageProducer is created and the resolved_sql
			// entry is sent to the lineage service.
			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 2",
				"resolved_sql": "SELECT 1 resolved again",
			}}

			cfg2 := asset.Config{ColumnLineageHost: srvSlow.URL, ColumnLineageChangeIdentifier: "data.optimus.resolved_sql"}
			updatedAsset2, producer2, err := r.repository.Upsert(r.ctx, &ast, false, cfg2)
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset2)
			r.Require().NotNil(producer2)

			// force a canceled context to deterministically trigger cancellation inside producer
			cancelledCtx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err = producer2(cancelledCtx)
			r.Require().Error(err)

			// unblock server to allow cleanup
			close(blockCh)
		})

		// Cover traversal of nested maps in changelog.To (multi-segment nestedKey)
		r.Run("should traverse nested maps in changelog.To and extract nested query", func() {
			// mock server to assert received query
			called := false
			var receivedQuery string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/api/v1/lineage/columns" {
					w.WriteHeader(404)
					return
				}
				var payload map[string]string
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					w.WriteHeader(400)
					return
				}
				q, ok := payload["query"]
				if !ok {
					w.WriteHeader(400)
					return
				}
				called = true
				receivedQuery = q
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"target_table": "p.s.t", "columns": []map[string]interface{}{}})
			}))
			defer srv.Close()

			// insert baseline asset without optimus to ensure changelog.Path is "data.optimus"
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-nested-traverse",
				Name:      "optimus-nested-traverse",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			// now update asset: set sql (so producer condition is met) and nested map for traversal
			nestedQuery := "SELECT nested FROM deep"
			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql": "SELECT 1",
				"some": map[string]interface{}{
					"path": map[string]interface{}{
						"query": nestedQuery,
					},
				},
			}}

			cfg := asset.Config{ColumnLineageHost: srv.URL, ColumnLineageChangeIdentifier: "data.optimus.some.path.query"}
			updatedAsset, producer, err := r.repository.Upsert(r.ctx, &ast, false, cfg)
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			r.Require().NotNil(producer)

			_, err = producer(context.Background())
			r.Require().NoError(err)
			r.True(called, "expected lineage service to be called")
			r.Equal(nestedQuery, receivedQuery)
		})

		r.Run("should return error when changelog.To is non-map but identifier expects nested keys", func() {
			// insert baseline asset with optimus as a map containing sql
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-nonmap-traverse",
				Name:      "optimus-nonmap-traverse",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			// update optimus to be a non-map (string) so changelog.To will be a string
			ast.Data = map[string]interface{}{"optimus": "not-a-map"}

			cfg := asset.Config{ColumnLineageHost: "http://example.invalid", ColumnLineageChangeIdentifier: "data.optimus.some.path.query"}
			updatedAsset, producer, err := r.repository.Upsert(r.ctx, &ast, false, cfg)
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			r.Require().NotNil(producer)

			// calling producer should return an error from extractColumnLineageQuery because it
			// expects a map while traversing nested keys
			_, err = producer(context.Background())
			r.Require().Error(err)
			r.ErrorContains(err, "expected map while traversing")
		})

		r.Run("should return empty graph when lineage response contains invalid entries", func() {
			// server returns invalid target_table or invalid source table
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(200)
				resp := map[string]interface{}{
					"target_table": "",
					"columns": []map[string]interface{}{{
						"target_column": "c",
						"sources":       []map[string]string{{"table": "", "column": "c1"}},
					}},
				}
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer srv.Close()

			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-optimus-invalid-response",
				Name:    "optimus-invalid-response",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{"optimus": map[string]interface{}{
					"sql": "SELECT 1",
				}},
				UpdatedBy: r.users[0],
			}

			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 1",
				"resolved_sql": "SELECT 1 resolved",
			}}

			cfg := asset.Config{ColumnLineageHost: srv.URL, ColumnLineageChangeIdentifier: "data.optimus.resolved_sql"}
			updatedAsset, producer, err := r.repository.Upsert(r.ctx, &ast, false, cfg)
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			r.Require().NotNil(producer)

			g, err := producer(context.Background())
			r.Require().NoError(err)
			// parser should produce empty graph for invalid entries
			r.Equal(0, len(g))
		})

		r.Run("should not set sql_version when optimus has no sql", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-init-no-sql",
				Name:      "optimus-init-no-sql",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"job_name": "some-job"}},
				UpdatedBy: r.users[0],
			}
			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)
			optimus := insertedAsset.Data["optimus"].(map[string]interface{})
			r.Nil(optimus["sql_version"])
			r.Nil(optimus["resolved_sql_version"])
		})
	})

	r.Run("on update", func() {
		r.Run("should not create nor updating the asset if asset is identical", func() {
			ast := asset.Asset{
				URN:         uuid.NewString() + "urn-u-2",
				Name:        "u-2",
				Type:        "table",
				Service:     "bigquery",
				UpdatedBy:   r.users[0],
				RefreshedAt: &refreshedAtTime,
				Version:     "0.1",
				Data:        map[string]interface{}{},
			}
			identicalAsset := ast

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)

			identicalAssetResult, _, err := r.repository.Upsert(r.ctx, &identicalAsset, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(identicalAssetResult.ID)

			r.Equal(insertedAsset.ID, identicalAssetResult.ID)
			r.Equal(identicalAssetResult.Version, identicalAssetResult.Version)
		})

		r.Run("should same asset version if asset only has different at RefreshedAt", func() {
			oneDayAgoRefreshedAtTime := refreshedAtTime.AddDate(0, 0, -1)
			ast := asset.Asset{
				URN:         uuid.NewString() + "urn-u-2",
				Name:        "u-2",
				Type:        "table",
				Service:     "bigquery",
				URL:         "https://sample-url-old.com",
				UpdatedBy:   r.users[0],
				RefreshedAt: &oneDayAgoRefreshedAtTime,
				Version:     "0.1",
				Data:        map[string]interface{}{},
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			ast.ID = insertedAsset.ID

			updated := ast
			updated.RefreshedAt = &refreshedAtTime

			upsertedAsset, _, err := r.repository.Upsert(r.ctx, &updated, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			r.Equal(insertedAsset.ID, upsertedAsset.ID)
			r.Equal(updated.RefreshedAt, upsertedAsset.RefreshedAt)
			r.Equal(ast.Version, upsertedAsset.Version)
		})

		r.Run("should update the asset version if asset is not identical", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-u-2",
				Name:      "u-2",
				Type:      "table",
				Service:   "bigquery",
				URL:       "https://sample-url-old.com",
				UpdatedBy: r.users[0],
				Version:   "0.1",
				Data:      map[string]interface{}{},
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			ast.ID = insertedAsset.ID

			updated := ast
			updated.URL = "https://sample-url.com"

			upsertedAsset, _, err := r.repository.Upsert(r.ctx, &updated, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			r.Equal(insertedAsset.ID, upsertedAsset.ID)
			r.Equal(updated.URL, upsertedAsset.URL)
			r.NotEqual(ast.Version, upsertedAsset.Version)
		})

		r.Run("should return not found error if the asset does not exist and isUpdateOnly is true", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-u-2",
				Type:      "table",
				Service:   "bigquery",
				URL:       "https://sample-url-old.com",
				UpdatedBy: r.users[0],
				Version:   "0.1",
			}

			upsertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, true, asset.Config{})
			r.Require().Error(err)
			r.ErrorIs(err, asset.NotFoundError{URN: ast.URN})
			r.Nil(upsertedAsset)
		})

		r.Run("should delete old owners if it does not exist on new asset", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-4",
				Name:    "u-4",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
					stripUserID(r.users[2]),
				},
				UpdatedBy: r.users[0],
				Data:      map[string]interface{}{},
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				stripUserID(r.users[2]),
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)

			upsertedAsset, _, err := r.repository.Upsert(r.ctx, &newAsset, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			r.Len(upsertedAsset.Owners, len(newAsset.Owners))
			r.Equal(r.users[2].ID, upsertedAsset.Owners[0].ID)
		})

		r.Run("should create new owners if it does not exist on old asset", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-4",
				Name:    "u-4",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				UpdatedBy: r.users[0],
				Data:      map[string]interface{}{},
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				stripUserID(r.users[1]),
				stripUserID(r.users[2]),
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)

			upsertedAsset, _, err := r.repository.Upsert(r.ctx, &newAsset, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			r.Len(upsertedAsset.Owners, len(newAsset.Owners))
			r.Equal(r.users[1].ID, upsertedAsset.Owners[0].ID)
			r.Equal(r.users[2].ID, upsertedAsset.Owners[1].ID)
		})

		r.Run("should create users from owners if owner emails do not exist yet", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-4a",
				Name:    "u-4a",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				UpdatedBy: r.users[0],
				Data:      map[string]interface{}{},
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				stripUserID(r.users[1]),
				{Email: "newuser@example.com"},
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)

			upsertedAsset, _, err := r.repository.Upsert(r.ctx, &newAsset, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			r.Len(upsertedAsset.Owners, len(newAsset.Owners))
			r.NotEmpty(upsertedAsset.Owners[0].ID)
			r.Equal(r.users[1].ID, upsertedAsset.Owners[0].ID)
			r.NotEmpty(upsertedAsset.Owners[1].ID)
			r.Equal(newAsset.Owners[1].Email, upsertedAsset.Owners[1].Email)
		})

		r.Run("should restore the asset and update the asset version if re-sync/update soft deleted asset", func() {
			updatedBy := r.users[0]
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-u-2",
				Name:      "u-2",
				Type:      "table",
				Service:   "bigquery",
				UpdatedBy: updatedBy,
				Version:   "0.1",
				Data:      map[string]interface{}{},
			}

			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			ast.ID = insertedAsset.ID

			newVersion, err := r.repository.SoftDeleteByURN(r.ctx, currentTime, ast.URN, updatedBy.ID)
			r.Require().NoError(err)
			r.Equal("0.2", newVersion) // successfully soft deleted the asset

			softDeletedAsset, err := r.repository.GetByURN(r.ctx, ast.URN)
			r.Require().NoError(err)
			r.Equal(true, softDeletedAsset.IsDeleted) // asset is soft deleted

			upsertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Equal(false, upsertedAsset.IsDeleted) // asset is restored
			r.Equal("0.3", upsertedAsset.Version)
		})

		r.Run("should bump sql_version when sql changes", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-bump-sql",
				Name:      "optimus-bump-sql",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 2"}}
			updatedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal("0.2", optimus["sql_version"])
		})

		r.Run("should not change sql_version when sql is unchanged", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-no-bump",
				Name:      "optimus-no-bump",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			ast.Description = "changed description only"
			updatedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
		})

		r.Run("should initialize resolved_sql_version when resolved_sql appears for the first time", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-resolved-first",
				Name:      "optimus-resolved-first",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 1",
				"resolved_sql": "SELECT 1 resolved",
			}}
			updatedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})

		r.Run("should not bump resolved_sql_version when resolved_sql already exists", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-optimus-resolved-no-bump",
				Name:    "optimus-resolved-no-bump",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{"optimus": map[string]interface{}{
					"sql":          "SELECT 1",
					"resolved_sql": "SELECT 1 resolved",
				}},
				UpdatedBy: r.users[0],
			}
			insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)

			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 1",
				"resolved_sql": "SELECT 1 resolved again",
			}}
			updatedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})

		r.Run("should bump sql_version and initialize resolved_sql_version when both change together", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-optimus-bump-both",
				Name:      "optimus-bump-both",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)

			ast.Data = map[string]interface{}{"optimus": map[string]interface{}{
				"sql":          "SELECT 2",
				"resolved_sql": "SELECT 2 resolved",
			}}
			updatedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal("0.2", optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})
	})
}

func (r *AssetRepositoryTestSuite) TestUpsertRaceCondition() {
	r.Run("TestUpsertRaceCondition", func() {
		ast := asset.Asset{
			URN:       "urn-upsert-race-condition",
			Name:      "upsert-race-condition",
			Type:      "table",
			Service:   "bigquery",
			URL:       "https://sample-url-old.com",
			UpdatedBy: r.users[0],
			Version:   "0.1",
			Data:      map[string]interface{}{},
		}

		insertedAsset, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)
		r.NotEmpty(insertedAsset.ID)

		const numGoroutines = 10 // Number of concurrent upserts
		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make([]error, 0, numGoroutines)

		// Concurrently upsert the object
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				localAst := ast
				localAst.URL = fmt.Sprintf("https://sample-url-%d.com", index)
				_, _, err := r.repository.Upsert(context.Background(), &localAst, false, asset.Config{})

				mu.Lock()
				results = append(results, err)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Check for errors
		for _, err := range results {
			assert.NoError(r.T(), err, "Upsert should not fail under race conditions")
		}
	})
}

func (r *AssetRepositoryTestSuite) TestUpsertPatch() {
	refreshedAtTime := time.Date(2024, time.August, 20, 8, 19, 49, 0, time.UTC)
	r.Run("on insert", func() {
		r.Run("should do nothing for patch data on insert action", func() {
			ast := asset.Asset{
				URN:     "urn-i-0",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					r.users[1],
				},
				Data: map[string]interface{}{
					"entity": "gojek",
					"data": map[string]interface{}{
						"foo": "bar",
					},
				},
				UpdatedBy: r.users[0],
			}

			patchData := make(map[string]interface{})
			patchData["data"] = map[string]interface{}{
				"entity": "gotocompany",
				"data": map[string]interface{}{
					"foo": "cookie",
				},
			}

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			r.NotEqual("gotocompany", insertedAsset.Data["entity"])
			r.Equal("gojek", insertedAsset.Data["entity"])
			r.Equal(map[string]interface{}{"foo": "bar"}, insertedAsset.Data["data"])
		})

		r.Run("should return not found error if the asset does not exist and isUpdateOnly is true", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-u-2",
				Type:      "table",
				Service:   "bigquery",
				URL:       "https://sample-url-old.com",
				UpdatedBy: r.users[0],
				Version:   "0.1",
			}

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, true, asset.Config{})
			r.Require().Error(err)
			r.ErrorIs(err, asset.NotFoundError{URN: ast.URN})
			r.Nil(upsertedAsset)
		})

		r.Run("set ID to asset and version to base version", func() {
			ast := asset.Asset{
				URN:     "urn-i-1",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				URL:     "https://sample-url.com",
				Data: map[string]interface{}{
					"entity": "gojek",
				},
				UpdatedBy:   r.users[0],
				RefreshedAt: &refreshedAtTime,
			}

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Equal(asset.BaseVersion, insertedAsset.Version)
			r.NoError(err)
			r.NotEmpty(insertedAsset.ID)
			r.NotEmpty(ast.CreatedAt)
			r.NotEmpty(ast.UpdatedAt)
			r.NotEqual(time.Time{}, insertedAsset.CreatedAt)
			r.NotEqual(time.Time{}, insertedAsset.UpdatedAt)

			ast.ID = insertedAsset.ID
			ast.Version = asset.BaseVersion
			r.assertAsset(&ast, insertedAsset)

			// Same with ast1
			ast2 := asset.Asset{
				URN:       ast.URN,
				Name:      ast.Name,
				Type:      ast.Type,
				Service:   ast.Service,
				URL:       ast.URL,
				Data:      make(map[string]interface{}),
				UpdatedBy: ast.UpdatedBy,
				Version:   ast.Version,
			}
			// Deep copy the Data map
			for key, value := range ast.Data {
				ast2.Data[key] = value
			}
			patchData := make(map[string]interface{})
			patchData["description"] = "create a new version" // to force fetch from asset_versions
			_, _, err = r.repository.UpsertPatch(r.ctx, &ast2, patchData, false, asset.Config{})
			r.NoError(err)
			assetv1, err := r.repository.GetByVersionWithID(r.ctx, ast.ID, asset.BaseVersion)
			r.NoError(err)
			r.Equal("0.1", assetv1.Version)
		})

		r.Run("should store owners if any", func() {
			ast := asset.Asset{
				URN:     "urn-i-3",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					r.users[1],
					{Email: r.users[2].Email},
					{ID: r.users[1].ID}, // should get deduplicated by ID
				},
				Data: map[string]interface{}{
					"entity": "gojek",
				},
				UpdatedBy: r.users[0],
			}

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotEmpty(insertedAsset.ID)
			r.Len(insertedAsset.Owners, 2)
			r.Equal(r.users[1].ID, insertedAsset.Owners[0].ID)
			r.Equal(r.users[2].ID, insertedAsset.Owners[1].ID)
		})

		r.Run("should create owners as users if they do not exist yet", func() {
			ast := asset.Asset{
				URN:     "urn-i-patch-insert-create-owner",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					{Email: "newuserpatch@example.com"},
					{Email: "newuserpatch2@example.com"},
					{Email: "newuserpatch@example.com"}, // should get deduplicated by ID on fetch user by email
				},
				Data: map[string]interface{}{
					"entity": "gojek",
				},
				UpdatedBy: r.users[0],
			}

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			r.Len(insertedAsset.Owners, 2)
			r.Equal(ast.Owners[0].Email, insertedAsset.Owners[0].Email)
		})

		r.Run("should initialize sql_version when sql is present", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-patch-optimus-init-sql",
				Name:      "optimus-init-sql",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)
			optimus := insertedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Nil(optimus["resolved_sql_version"])
		})

		r.Run("should initialize both versions when sql and resolved_sql are present", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-patch-optimus-init-both",
				Name:    "optimus-init-both",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{"optimus": map[string]interface{}{
					"sql":          "SELECT 1",
					"resolved_sql": "SELECT 2",
				}},
				UpdatedBy: r.users[0],
			}
			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)
			optimus := insertedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})

		r.Run("should not set sql_version when optimus has no sql", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-patch-optimus-init-no-sql",
				Name:      "optimus-init-no-sql",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"job_name": "some-job"}},
				UpdatedBy: r.users[0],
			}
			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(insertedAsset)
			optimus := insertedAsset.Data["optimus"].(map[string]interface{})
			r.Nil(optimus["sql_version"])
			r.Nil(optimus["resolved_sql_version"])
		})
	})

	r.Run("on update", func() {
		r.Run("should not create nor updating the asset if asset is identical", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-2",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{
					"entity": "gotocompany",
				},
				UpdatedBy:   r.users[0],
				RefreshedAt: &refreshedAtTime,
				Version:     "0.1",
			}
			identicalAsset := ast

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{}) // insert
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)

			patchData := make(map[string]interface{})
			patchData["data"] = map[string]interface{}{
				"entity": "gotocompany",
			}

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &identicalAsset, patchData, false, asset.Config{}) // update
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)

			r.Equal(insertedAsset.ID, upsertedAsset.ID)
			r.Equal(insertedAsset.Version, upsertedAsset.Version)
			r.Equal(upsertedAsset.Data["entity"], "gotocompany")
		})

		r.Run("should same asset version if asset only has different at RefreshedAt", func() {
			oneDayAgoRefreshedAtTime := refreshedAtTime.AddDate(0, 0, -1)
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-2",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				URL:     "https://sample-url-old.com",
				Data: map[string]interface{}{
					"entity": "gotocompany",
				},
				UpdatedBy:   r.users[0],
				RefreshedAt: &oneDayAgoRefreshedAtTime,
				Version:     "0.1",
			}

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{}) // insert
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			ast.ID = insertedAsset.ID

			updated := ast
			updated.RefreshedAt = &refreshedAtTime
			patchData := make(map[string]interface{})
			patchData["data"] = map[string]interface{}{
				"entity": "gotocompany",
			}

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &updated, patchData, false, asset.Config{}) // update
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			updated.ID = upsertedAsset.ID

			r.Equal(insertedAsset.ID, upsertedAsset.ID)
			r.Equal(updated.RefreshedAt, upsertedAsset.RefreshedAt)
			r.Equal(insertedAsset.Version, upsertedAsset.Version)
			r.Equal(upsertedAsset.Data["entity"], "gotocompany")
		})

		r.Run("should update the asset version if asset is not identical", func() {
			ast := asset.Asset{
				URN:         uuid.NewString() + "urn-u-2",
				Description: "existing",
				Name:        "urn-test",
				Type:        "table",
				Service:     "bigquery",
				URL:         "https://sample-url.com",
				Data: map[string]interface{}{
					"entity": "gotocompany",
					"data": map[string]interface{}{
						"foo": "old",
					},
				},
				UpdatedBy: r.users[0],
				Version:   "0.1",
			}

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{}) // insert
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			ast.ID = insertedAsset.ID

			updated := ast
			updated.Description = "bluffing"
			patchData := map[string]interface{}{}
			patchData["data"] = map[string]interface{}{
				"data": map[string]interface{}{
					"foo": "new",
				},
			}
			updated.Patch(patchData)

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &updated, patchData, false, asset.Config{}) // update
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			updated.ID = upsertedAsset.ID

			r.Equal(insertedAsset.ID, upsertedAsset.ID)
			r.Equal("existing", upsertedAsset.Description)
			r.Equal("gotocompany", upsertedAsset.Data["entity"])
			r.Equal(map[string]interface{}{"foo": "new"}, upsertedAsset.Data["data"])
			r.NotEqual(ast.Version, upsertedAsset.Version)
		})

		r.Run("should keep old data if it does not exist on new asset", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-4",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				Data: map[string]interface{}{
					"entity": "gotocompany",
				},
				UpdatedBy: r.users[0],
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				stripUserID(r.users[2]),
			}

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{}) // insert
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)

			patchData := make(map[string]interface{})
			patchData["data"] = map[string]interface{}{
				"another": "things",
			}

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &newAsset, patchData, false, asset.Config{}) // update
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			r.Len(upsertedAsset.Data, 2)
			r.Equal(upsertedAsset.Data["entity"], "gotocompany")
			r.Equal(upsertedAsset.Data["another"], "things")
		})

		r.Run("should delete old owners if it does not exist on new patch", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-4",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				Data: map[string]interface{}{
					"entity": "gotocompany",
				},
				UpdatedBy: r.users[0],
			}
			newAsset := ast

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)
			ast.ID = insertedAsset.ID

			patchData := make(map[string]interface{})
			patchData["owners"] = []map[string]interface{}{
				{
					"id":       r.users[2].ID,
					"email":    r.users[2].Email,
					"provider": r.users[2].Provider,
				},
			}

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &newAsset, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			r.Len(upsertedAsset.Owners, len(newAsset.Owners))
			r.Equal(r.users[2].ID, upsertedAsset.Owners[0].ID)
		})

		r.Run("should create new owners if it does not exist on old asset", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-4",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				Data: map[string]interface{}{
					"entity": "gotocompany",
				},
				UpdatedBy: r.users[0],
			}
			newAsset := ast

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset.ID)

			patchData := make(map[string]interface{})
			patchData["owners"] = []map[string]interface{}{
				{
					"id":       r.users[1].ID,
					"email":    r.users[1].Email,
					"provider": r.users[1].Provider,
				},
				{
					"id":       r.users[2].ID,
					"email":    r.users[2].Email,
					"provider": r.users[2].Provider,
				},
			}

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &newAsset, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset.ID)
			r.Len(upsertedAsset.Owners, 2)
			r.Equal(r.users[1].ID, upsertedAsset.Owners[0].ID)
			r.Equal(r.users[2].ID, upsertedAsset.Owners[1].ID)
		})

		r.Run("should create users from owners if owner emails do not exist yet", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-u-4a",
				Name:    "urn-test",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				Data: map[string]interface{}{
					"entity": "gotocompany",
				},
				UpdatedBy: r.users[0],
			}
			newAsset := ast

			insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{}) // insert
			r.Require().NoError(err)
			r.NotEmpty(insertedAsset)

			patchData := make(map[string]interface{})
			patchData["owners"] = []map[string]interface{}{
				{
					"id":       r.users[1].ID,
					"email":    r.users[1].Email,
					"provider": r.users[1].Provider,
				},
				{
					"email": "newuser@example.com",
				},
			}

			upsertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &newAsset, patchData, false, asset.Config{}) // update
			r.Require().NoError(err)
			r.NotEmpty(upsertedAsset)
			r.Len(upsertedAsset.Owners, 2)
			r.NotEmpty(upsertedAsset.Owners[0].ID)
			r.Equal(r.users[1].ID, upsertedAsset.Owners[0].ID)
			r.NotEmpty(upsertedAsset.Owners[1].ID)
		})

		r.Run("should bump sql_version when sql changes", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-patch-optimus-bump-sql",
				Name:      "optimus-bump-sql",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)

			patchData := map[string]interface{}{
				"data": map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 2"}},
			}
			updatedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal("0.2", optimus["sql_version"])
		})

		r.Run("should not change sql_version when sql is unchanged", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-patch-optimus-no-bump",
				Name:      "optimus-no-bump",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)

			patchData := map[string]interface{}{"description": "only description changed"}
			updatedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
		})

		r.Run("should initialize resolved_sql_version when resolved_sql appears for the first time", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-patch-optimus-resolved-first",
				Name:      "optimus-resolved-first",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)

			patchData := map[string]interface{}{
				"data": map[string]interface{}{"optimus": map[string]interface{}{
					"sql":          "SELECT 1",
					"resolved_sql": "SELECT 1 resolved",
				}},
			}
			updatedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})

		r.Run("should not bump resolved_sql_version when resolved_sql already exists", func() {
			ast := asset.Asset{
				URN:     uuid.NewString() + "urn-patch-optimus-resolved-no-bump",
				Name:    "optimus-resolved-no-bump",
				Type:    "table",
				Service: "bigquery",
				Data: map[string]interface{}{"optimus": map[string]interface{}{
					"sql":          "SELECT 1",
					"resolved_sql": "SELECT 1 resolved",
				}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)

			patchData := map[string]interface{}{
				"data": map[string]interface{}{"optimus": map[string]interface{}{
					"sql":          "SELECT 1",
					"resolved_sql": "SELECT 1 resolved again",
				}},
			}
			updatedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal(asset.BaseVersion, optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})

		r.Run("should bump sql_version and initialize resolved_sql_version when both change together", func() {
			ast := asset.Asset{
				URN:       uuid.NewString() + "urn-patch-optimus-bump-both",
				Name:      "optimus-bump-both",
				Type:      "table",
				Service:   "bigquery",
				Data:      map[string]interface{}{"optimus": map[string]interface{}{"sql": "SELECT 1"}},
				UpdatedBy: r.users[0],
			}
			_, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
			r.Require().NoError(err)

			patchData := map[string]interface{}{
				"data": map[string]interface{}{"optimus": map[string]interface{}{
					"sql":          "SELECT 2",
					"resolved_sql": "SELECT 2 resolved",
				}},
			}
			updatedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, patchData, false, asset.Config{})
			r.Require().NoError(err)
			r.Require().NotNil(updatedAsset)
			optimus := updatedAsset.Data["optimus"].(map[string]interface{})
			r.Equal("0.2", optimus["sql_version"])
			r.Equal(asset.BaseVersion, optimus["resolved_sql_version"])
		})
	})
}

func (r *AssetRepositoryTestSuite) TestUpsertPatchRaceCondition() {
	r.Run("TestUpsertPatchRaceCondition", func() {
		ast := asset.Asset{
			URN:     "urn-upsert-patch-race-condition",
			Name:    "urn-test",
			Type:    "table",
			Service: "bigquery",
			URL:     "https://sample-url-old.com",
			Data: map[string]interface{}{
				"entity": "gotocompany",
			},
			UpdatedBy: r.users[0],
			Version:   "0.1",
		}

		insertedAsset, _, err := r.repository.UpsertPatch(r.ctx, &ast, nil, false, asset.Config{})
		r.Require().NoError(err)
		r.NotEmpty(insertedAsset.ID)

		const numGoroutines = 10 // Number of concurrent upserts
		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make([]error, 0, numGoroutines)

		// Concurrently upsert the object
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				localAst := ast
				patchData := make(map[string]interface{})
				patchData["data"] = map[string]interface{}{
					"entity": fmt.Sprintf("entity-%d", index),
				}
				_, _, err := r.repository.UpsertPatch(context.Background(), &localAst, patchData, false, asset.Config{})

				mu.Lock()
				results = append(results, err)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Check for errors
		for _, err := range results {
			assert.NoError(r.T(), err, "Upsert should not fail under race conditions")
		}
	})
}

func (r *AssetRepositoryTestSuite) TestDeleteByID() {
	r.Run("return error from client if any", func() {
		_, err := r.repository.DeleteByID(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
	})

	r.Run("return NotFoundError if asset does not exist", func() {
		uuidTest := "2aabb450-f986-44e2-a6db-7996861d5004"
		_, err := r.repository.DeleteByID(r.ctx, uuidTest)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuidTest})
	})

	r.Run("should delete correct asset", func() {
		asset1 := asset.Asset{
			URN:       "urn-del-1",
			Name:      "del-1",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:       "urn-del-2",
			Name:      "del-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}

		var err error
		insertedAsset, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(insertedAsset.ID)
		asset1.ID = insertedAsset.ID

		insertedAsset2, _, err := r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(insertedAsset2.ID)
		asset2.ID = insertedAsset2.ID

		_, err = r.repository.DeleteByID(r.ctx, asset1.ID)
		r.NoError(err)

		_, err = r.repository.GetByID(r.ctx, asset1.ID)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: asset1.ID})

		asset2FromDB, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		r.Equal(asset2.ID, asset2FromDB.ID)

		// cleanup
		_, err = r.repository.DeleteByID(r.ctx, asset2.ID)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestSoftDeleteByID() {
	currentTime := time.Now().UTC()
	userID := r.users[0].ID

	r.Run("return NotFoundError if asset does not exist", func() {
		uuidTest := "2aabb450-f986-44e2-a6db-7996861d5004"
		_, _, err := r.repository.SoftDeleteByID(r.ctx, currentTime, uuidTest, userID)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuidTest})
	})

	r.Run("should delete correct asset", func() {
		asset1 := asset.Asset{
			URN:       "urn-del-1",
			Name:      "del-1",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: user.User{ID: userID},
			Data:      map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:       "urn-del-2",
			Name:      "del-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: user.User{ID: userID},
			Data:      map[string]interface{}{},
		}

		var err error
		ast1, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(ast1.ID)
		asset1.ID = ast1.ID

		ast2, _, err := r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)
		r.Require().NotEmpty(ast2.ID)
		asset2.ID = ast2.ID

		_, _, err = r.repository.SoftDeleteByID(r.ctx, currentTime, asset1.ID, userID)
		r.NoError(err)

		asset1FromDB, err := r.repository.GetByID(r.ctx, asset1.ID)
		r.Require().NoError(err)
		r.True(asset1FromDB.IsDeleted)

		asset2FromDB, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		r.False(asset2FromDB.IsDeleted)

		// cleanup
		_, err = r.repository.DeleteByID(r.ctx, asset1.ID)
		r.NoError(err)
		_, err = r.repository.DeleteByID(r.ctx, asset2.ID)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestDeleteByURN() {
	r.Run("return NotFoundError if asset does not exist", func() {
		urn := "urn-test-1"
		err := r.repository.DeleteByURN(r.ctx, urn)
		r.ErrorContains(err, "query affected 0 rows")
	})

	r.Run("should delete correct asset", func() {
		asset1 := asset.Asset{
			URN:       "urn-del-1",
			Name:      "del-1",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:       "urn-del-2",
			Name:      "del-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}

		_, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)

		insertedAsset2, _, err := r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)

		err = r.repository.DeleteByURN(r.ctx, asset1.URN)
		r.NoError(err)

		_, err = r.repository.GetByURN(r.ctx, asset1.URN)
		r.ErrorAs(err, &asset.NotFoundError{URN: asset1.URN})

		asset2FromDB, err := r.repository.GetByURN(r.ctx, asset2.URN)
		r.NoError(err)
		r.Equal(insertedAsset2.ID, asset2FromDB.ID)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, asset2.URN)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestSoftDeleteByURN() {
	currentTime := time.Now().UTC()
	userID := r.users[0].ID

	r.Run("return NotFoundError if asset does not exist", func() {
		urn := "urn-test-1"
		_, err := r.repository.SoftDeleteByURN(r.ctx, currentTime, urn, userID)
		r.ErrorContainsf(err, "could not find asset", "urn = %s", urn)
	})

	r.Run("should delete correct asset", func() {
		asset1 := asset.Asset{
			URN:       "urn-del-1",
			Name:      "del-1",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: user.User{ID: userID},
			Data:      map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:       "urn-del-2",
			Name:      "del-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: user.User{ID: userID},
			Data:      map[string]interface{}{},
		}

		_, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)

		_, _, err = r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)

		_, err = r.repository.SoftDeleteByURN(r.ctx, currentTime, asset1.URN, userID)
		r.NoError(err)

		asset1FromDB, err := r.repository.GetByURN(r.ctx, asset1.URN)
		r.Require().NoError(err)
		r.True(asset1FromDB.IsDeleted)

		asset2FromDB, err := r.repository.GetByURN(r.ctx, asset2.URN)
		r.NoError(err)
		r.False(asset2FromDB.IsDeleted)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, asset1.URN)
		r.NoError(err)
		err = r.repository.DeleteByURN(r.ctx, asset2.URN)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestDeleteByIsDeletedAndServicesAndUpdatedAt() {
	currentTime := time.Now()
	serviceA := "serviceA"
	serviceB := "serviceB"

	// Insert assets with different services and updated_at
	asset1 := asset.Asset{
		URN:       "urn-del-service-1",
		Name:      "Asset 1",
		Service:   serviceA,
		Type:      "topic",
		IsDeleted: false,
		UpdatedBy: r.users[0],
		Data:      map[string]interface{}{},
	}
	asset2 := asset.Asset{
		URN:       "urn-del-service-2",
		Name:      "Asset 2",
		Service:   serviceB,
		Type:      "topic",
		IsDeleted: true,
		UpdatedBy: r.users[0],
		Data:      map[string]interface{}{},
	}
	asset3 := asset.Asset{
		URN:       "urn-del-service-3",
		Name:      "Asset 3",
		Service:   serviceA,
		Type:      "topic",
		IsDeleted: true,
		UpdatedBy: r.users[0],
		Data:      map[string]interface{}{},
	}

	_, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
	r.Require().NoError(err)
	_, _, err = r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
	r.Require().NoError(err)
	_, _, err = r.repository.Upsert(r.ctx, &asset3, false, asset.Config{})
	r.Require().NoError(err)

	thresholdUpdatedAt := currentTime.Add(1 * time.Hour)

	r.Run("should return error if services is empty", func() {
		urns, err := r.repository.DeleteByIsDeletedAndServicesAndUpdatedAt(r.ctx, true, []string{}, thresholdUpdatedAt)
		r.Nil(urns)
		r.ErrorIs(err, asset.ErrEmptyServices)
	})

	r.Run("should delete assets for given is_deleted, services, and updated_at", func() {
		urns, err := r.repository.DeleteByIsDeletedAndServicesAndUpdatedAt(r.ctx, false, []string{serviceA}, thresholdUpdatedAt)
		r.NoError(err)
		r.ElementsMatch([]string{"urn-del-service-1"}, urns)

		// asset1 should be deleted, asset3 should remain, asset2 should remain
		_, err = r.repository.GetByURN(r.ctx, asset1.URN)
		r.ErrorAs(err, &asset.NotFoundError{URN: asset1.URN})

		_, err = r.repository.GetByURN(r.ctx, asset3.URN)
		r.NoError(err)
		_, err = r.repository.GetByURN(r.ctx, asset2.URN)
		r.NoError(err)
	})

	r.Run("should delete all assets if AllServicesCleanupConfig is set", func() {
		urns, err := r.repository.DeleteByIsDeletedAndServicesAndUpdatedAt(r.ctx, true, []string{asset.AllServicesCleanupConfig}, thresholdUpdatedAt)
		r.NoError(err)
		r.ElementsMatch([]string{"urn-del-service-2", "urn-del-service-3"}, urns)

		_, err = r.repository.GetByURN(r.ctx, asset2.URN)
		r.ErrorAs(err, &asset.NotFoundError{URN: asset2.URN})

		_, err = r.repository.GetByURN(r.ctx, asset3.URN)
		r.ErrorAs(err, &asset.NotFoundError{URN: asset2.URN})
	})

	// cleanup done by the test cases
}

func (r *AssetRepositoryTestSuite) TestDeleteByQueryExpr() {
	refreshedAtTime := time.Date(2024, time.August, 20, 8, 19, 49, 0, time.UTC)
	r.Run("should delete correct asset", func() {
		oneYearAgoRefreshedAtTime := refreshedAtTime.AddDate(-1, 0, 0)
		asset1 := asset.Asset{
			URN:         "urn-del-1",
			Name:        "del-1",
			Type:        "table",
			Service:     "bigquery",
			UpdatedBy:   r.users[0],
			RefreshedAt: &oneYearAgoRefreshedAtTime,
			Data:        map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:         "urn-del-2",
			Name:        "del-2",
			Type:        "topic",
			Service:     "kafka",
			Version:     asset.BaseVersion,
			UpdatedBy:   r.users[0],
			RefreshedAt: &oneYearAgoRefreshedAtTime,
			Data:        map[string]interface{}{},
		}

		_, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)

		insertedAsset2, _, err := r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)

		query := "refreshed_at <= '" + refreshedAtTime.Format("2006-01-02T15:04:05Z") +
			"' && service == '" + asset1.Service +
			"' && type == '" + asset1.Type.String() +
			"' && urn == '" + asset1.URN + "'"
		sqlExpr := queryexpr.SQLExpr(query)
		queryExpr := asset.DeleteAssetExpr{
			ExprStr: sqlExpr,
		}
		urns, err := r.repository.DeleteByQueryExpr(r.ctx, queryExpr)
		r.NoError(err)
		r.Equal([]string{"urn-del-1"}, urns)

		_, err = r.repository.GetByURN(r.ctx, asset1.URN)
		r.ErrorAs(err, &asset.NotFoundError{URN: asset1.URN})

		asset2FromDB, err := r.repository.GetByURN(r.ctx, asset2.URN)
		r.NoError(err)
		r.Equal(insertedAsset2.ID, asset2FromDB.ID)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, asset2.URN)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestSoftDeleteByQueryExpr() {
	refreshedAtTime := time.Date(2024, time.August, 20, 8, 19, 49, 0, time.UTC)
	currentTime := time.Now().UTC()
	r.Run("should delete correct asset", func() {
		userID := r.users[0].ID
		oneYearAgoRefreshedAtTime := refreshedAtTime.AddDate(-1, 0, 0)
		asset1 := asset.Asset{
			URN:         "urn-del-1",
			Name:        "del-1",
			Type:        "table",
			Service:     "bigquery",
			UpdatedBy:   user.User{ID: userID},
			RefreshedAt: &oneYearAgoRefreshedAtTime,
			Data:        map[string]interface{}{},
		}
		asset2 := asset.Asset{
			URN:         "urn-del-2",
			Name:        "del-2",
			Type:        "topic",
			Service:     "kafka",
			Version:     asset.BaseVersion,
			UpdatedBy:   user.User{ID: userID},
			RefreshedAt: &oneYearAgoRefreshedAtTime,
			Data:        map[string]interface{}{},
		}

		_, _, err := r.repository.Upsert(r.ctx, &asset1, false, asset.Config{})
		r.Require().NoError(err)

		_, _, err = r.repository.Upsert(r.ctx, &asset2, false, asset.Config{})
		r.Require().NoError(err)

		query := "refreshed_at <= '" + refreshedAtTime.Format("2006-01-02T15:04:05Z") +
			"' && service == '" + asset1.Service +
			"' && type == '" + asset1.Type.String() +
			"' && urn == '" + asset1.URN + "'"
		sqlExpr := queryexpr.SQLExpr(query)
		queryExpr := asset.DeleteAssetExpr{
			ExprStr: sqlExpr,
		}

		_, err = r.repository.SoftDeleteByQueryExpr(r.ctx, currentTime, userID, queryExpr)
		r.NoError(err)

		asset1FromDB, err := r.repository.GetByURN(r.ctx, asset1.URN)
		r.NoError(err)
		r.True(asset1FromDB.IsDeleted)

		asset2FromDB, err := r.repository.GetByURN(r.ctx, asset2.URN)
		r.NoError(err)
		r.False(asset2FromDB.IsDeleted)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, asset1.URN)
		r.NoError(err)
		err = r.repository.DeleteByURN(r.ctx, asset2.URN)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestAddProbe() {
	const typeJob = asset.Type("job")

	r.Run("return NotFoundError if asset does not exist", func() {
		urn := "invalid-urn"
		probe := asset.Probe{}
		err := r.repository.AddProbe(r.ctx, urn, &probe)
		r.ErrorAs(err, &asset.NotFoundError{URN: urn})
	})

	r.Run("should return error if probe already exists", func() {
		ast := asset.Asset{
			URN:       "urn-add-probe-1",
			Name:      "add-probe-1",
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		probeID := uuid.NewString()
		probe := asset.Probe{
			ID:           probeID,
			Status:       "COMPLETED",
			StatusReason: "Sample Reason",
			Timestamp:    time.Now().Add(2 * time.Minute),
			Metadata: map[string]interface{}{
				"foo": "bar",
			},
		}

		_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)

		err = r.repository.AddProbe(r.ctx, ast.URN, &probe)
		r.NoError(err)

		err = r.repository.AddProbe(r.ctx, ast.URN, &probe)
		r.ErrorIs(err, asset.ErrProbeExists)
	})

	r.Run("should populate CreatedAt and persist probe", func() {
		r.BeforeTest("", "")
		ast := asset.Asset{
			URN:       "urn-add-probe-1",
			Name:      "add-probe-1",
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		probe := asset.Probe{
			Status:       "COMPLETED",
			StatusReason: "Sample Reason",
			Timestamp:    time.Now().Add(2 * time.Minute),
			Metadata: map[string]interface{}{
				"foo": "bar",
			},
		}

		_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)

		err = r.repository.AddProbe(r.ctx, ast.URN, &probe)
		r.NoError(err)

		// assert populated fields
		r.NotEmpty(probe.ID)
		r.Equal(ast.URN, probe.AssetURN)
		r.False(probe.CreatedAt.IsZero())

		// assert probe is persisted
		probesFromDB, err := r.repository.GetProbes(r.ctx, ast.URN)
		r.Require().NoError(err)
		r.Require().Len(probesFromDB, 1)

		probeFromDB := probesFromDB[0]
		r.Equal(probe.ID, probeFromDB.ID)
		r.Equal(probe.AssetURN, probeFromDB.AssetURN)
		r.Equal(probe.Status, probeFromDB.Status)
		r.Equal(probe.StatusReason, probeFromDB.StatusReason)
		r.Equal(probe.Metadata, probeFromDB.Metadata)
		// we use `1µs` instead of `0` for the delta due to postgres might round precision before storing
		r.WithinDuration(probe.Timestamp, probeFromDB.Timestamp, 1*time.Microsecond)
		r.WithinDuration(probe.CreatedAt, probeFromDB.CreatedAt, 1*time.Microsecond)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, ast.URN)
		r.Require().NoError(err)
	})

	r.Run("should insert ID if specified", func() {
		ast := asset.Asset{
			URN:       "urn-add-probe-1",
			Name:      "add-probe-1",
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		probeID := uuid.NewString()
		probe := asset.Probe{
			ID:           probeID,
			Status:       "COMPLETED",
			StatusReason: "Sample Reason",
			Timestamp:    time.Now().Add(2 * time.Minute),
			Metadata: map[string]interface{}{
				"foo": "bar",
			},
		}

		_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)

		err = r.repository.AddProbe(r.ctx, ast.URN, &probe)
		r.NoError(err)

		r.Equal(probeID, probe.ID)
	})

	r.Run("should populate Timestamp if empty", func() {
		ast := asset.Asset{
			URN:       "urn-add-probe-2",
			Name:      "add-probe-2",
			Type:      typeJob,
			Service:   "optimus",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		otherAst := asset.Asset{
			URN:       "urn-add-probe-3",
			Name:      "add-probe-3",
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		probe := asset.Probe{
			Status: "RUNNING",
		}
		otherProbe := asset.Probe{
			Status: "RUNNING",
		}

		_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)
		_, _, err = r.repository.Upsert(r.ctx, &otherAst, false, asset.Config{})
		r.Require().NoError(err)

		err = r.repository.AddProbe(r.ctx, ast.URN, &probe)
		r.NoError(err)

		// assert populated fields
		r.False(probe.CreatedAt.IsZero())
		r.Equal(probe.CreatedAt, probe.Timestamp)

		err = r.repository.AddProbe(r.ctx, otherAst.URN, &otherProbe)
		r.NoError(err)

		// assert probe is persisted
		probesFromDB, err := r.repository.GetProbes(r.ctx, ast.URN)
		r.Require().NoError(err)
		r.Require().Len(probesFromDB, 1)

		probeFromDB := probesFromDB[0]
		r.Equal(probe.ID, probeFromDB.ID)
		r.WithinDuration(probe.Timestamp, probeFromDB.Timestamp, 1*time.Microsecond)
		r.WithinDuration(probe.CreatedAt, probeFromDB.CreatedAt, 1*time.Microsecond)
		r.WithinDuration(probeFromDB.CreatedAt, probeFromDB.Timestamp, 1*time.Microsecond)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, ast.URN)
		r.Require().NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestGetProbes() {
	r.Run("should return list of probes by asset urn", func() {
		ast := asset.Asset{
			URN:       "urn-add-probe-1",
			Name:      "add-probe-1",
			Type:      asset.Type("job"),
			Service:   "airflow",
			UpdatedBy: r.users[0],
			Data:      map[string]interface{}{},
		}
		p1 := asset.Probe{
			Status:    "COMPLETED",
			Timestamp: time.Now().UTC().Add(3 * time.Minute),
			Metadata: map[string]interface{}{
				"foo": "bar",
			},
		}
		p2 := asset.Probe{
			Status:       "FAILED",
			StatusReason: "sample error",
			Metadata: map[string]interface{}{
				"bar": "foo",
			},
		}
		p3 := asset.Probe{
			Status: "RUNNING",
		}

		_, _, err := r.repository.Upsert(r.ctx, &ast, false, asset.Config{})
		r.Require().NoError(err)

		err = r.repository.AddProbe(r.ctx, ast.URN, &p1)
		r.NoError(err)
		err = r.repository.AddProbe(r.ctx, ast.URN, &p2)
		r.NoError(err)
		err = r.repository.AddProbe(r.ctx, ast.URN, &p3)
		r.NoError(err)

		// assert probe is persisted
		actual, err := r.repository.GetProbes(r.ctx, ast.URN)
		r.Require().NoError(err)
		r.Require().Len(actual, 3)

		expected := []asset.Probe{p2, p3, p1}
		r.Equal(expected[0].ID, actual[0].ID)
		r.Equal(expected[1].ID, actual[1].ID)
		r.Equal(expected[2].ID, actual[2].ID)

		r.Equal(expected[0].ID, actual[0].ID)
		r.Equal(expected[0].AssetURN, actual[0].AssetURN)
		r.Equal(expected[0].Status, actual[0].Status)
		r.Equal(expected[0].StatusReason, actual[0].StatusReason)
		r.Equal(expected[0].Metadata, actual[0].Metadata)
		r.WithinDuration(expected[0].Timestamp, actual[0].Timestamp, 1*time.Microsecond)
		r.WithinDuration(expected[0].CreatedAt, actual[0].CreatedAt, 1*time.Microsecond)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, ast.URN)
		r.Require().NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestGetProbesWithFilter() {
	r.insertProbes(r.T())

	newTS := func(s string) time.Time {
		r.T().Helper()

		ts, err := time.Parse(time.RFC3339, s)
		r.Require().NoError(err)
		return ts
	}
	keys := func(m map[string][]asset.Probe) []string {
		kk := make([]string, 0, len(m))
		for k := range m {
			kk = append(kk, k)
		}
		return kk
	}

	cases := []struct {
		name     string
		flt      asset.ProbesFilter
		expected map[string][]asset.Probe
	}{
		{
			name: "AssetURNs=c-demo-kafka",
			flt:  asset.ProbesFilter{AssetURNs: []string{"c-demo-kafka"}},
			expected: map[string][]asset.Probe{
				"c-demo-kafka": {
					{
						AssetURN:  "c-demo-kafka",
						Status:    "SUCCESS",
						Timestamp: newTS("2022-03-08T09:58:43Z"),
					},
					{
						AssetURN:  "c-demo-kafka",
						Status:    "FAILURE",
						Timestamp: newTS("2021-11-25T19:28:18Z"),
					},
					{
						AssetURN:     "c-demo-kafka",
						Status:       "FAILURE",
						StatusReason: "Expanded even-keeled data-warehouse",
						Timestamp:    newTS("2021-11-10T09:28:21Z"),
					},
				},
			},
		},
		{
			name: "NewerThan=2022-09-08",
			flt:  asset.ProbesFilter{NewerThan: newTS("2022-09-08T00:00:00Z")},
			expected: map[string][]asset.Probe{
				"f-john-test-001": {
					{
						AssetURN:  "f-john-test-001",
						Status:    "CANCELLED",
						Timestamp: newTS("2022-09-23T14:39:57Z"),
					},
				},
				"ten-mock": {
					{
						AssetURN:     "ten-mock",
						Status:       "CANCELLED",
						StatusReason: "Synergized bottom-line forecast",
						Timestamp:    newTS("2022-09-11T07:40:11Z"),
					},
				},
				"e-test-grant2": {
					{
						AssetURN:     "e-test-grant2",
						Status:       "TERMINATED",
						StatusReason: "Digitized asynchronous knowledge user",
						Timestamp:    newTS("2022-09-08T12:16:42Z"),
					},
				},
			},
		},
		{
			name: "OlderThan=2021-11-01",
			flt:  asset.ProbesFilter{OlderThan: newTS("2021-11-01T00:00:00Z")},
			expected: map[string][]asset.Probe{
				"i-undefined-dfgdgd-avi": {
					{
						AssetURN:     "i-undefined-dfgdgd-avi",
						Status:       "CANCELLED",
						StatusReason: "Re-contextualized secondary projection",
						Timestamp:    newTS("2021-10-17T19:14:51Z"),
					},
				},
				"e-test-grant2": {
					{
						AssetURN:     "e-test-grant2",
						Status:       "TERMINATED",
						StatusReason: "Integrated attitude-oriented open system",
						Timestamp:    newTS("2021-10-31T05:58:13Z"),
					},
				},
			},
		},
		{
			name: "MaxRows=1",
			flt:  asset.ProbesFilter{MaxRows: 1},
			expected: map[string][]asset.Probe{
				"c-demo-kafka": {
					{
						AssetURN:  "c-demo-kafka",
						Status:    "SUCCESS",
						Timestamp: newTS("2022-03-08T09:58:43Z"),
					},
				},
				"e-test-grant2": {
					{
						AssetURN:     "e-test-grant2",
						Status:       "TERMINATED",
						StatusReason: "Digitized asynchronous knowledge user",
						Timestamp:    newTS("2022-09-08T12:16:42Z"),
					},
				},
				"eleven-mock": {
					{
						AssetURN:     "eleven-mock",
						Status:       "FAILURE",
						StatusReason: "Proactive zero administration attitude",
						Timestamp:    newTS("2022-02-21T22:52:06Z"),
					},
				},
				"f-john-test-001": {
					{
						AssetURN:  "f-john-test-001",
						Status:    "CANCELLED",
						Timestamp: newTS("2022-09-23T14:39:57Z"),
					},
				},
				"g-jane-kafka-1a": {
					{
						AssetURN:     "g-jane-kafka-1a",
						Status:       "SUCCESS",
						StatusReason: "Integrated 24/7 knowledge base",
						Timestamp:    newTS("2022-04-19T19:42:09Z"),
					},
				},
				"h-test-new-kafka": {
					{
						AssetURN:     "h-test-new-kafka",
						Status:       "SUCCESS",
						StatusReason: "User-friendly systematic neural-net",
						Timestamp:    newTS("2022-08-14T03:04:44Z"),
					},
				},
				"i-test-grant": {
					{
						AssetURN:     "i-test-grant",
						Status:       "FAILURE",
						StatusReason: "Ameliorated explicit customer loyalty",
						Timestamp:    newTS("2022-07-24T06:52:27Z"),
					},
				},
				"i-undefined-dfgdgd-avi": {
					{
						AssetURN:     "i-undefined-dfgdgd-avi",
						Status:       "TERMINATED",
						StatusReason: "Networked analyzing framework",
						Timestamp:    newTS("2022-08-13T13:54:01Z"),
					},
				},
				"j-xcvcx": {
					{
						AssetURN:     "j-xcvcx",
						Status:       "FAILURE",
						StatusReason: "Compatible impactful workforce",
						Timestamp:    newTS("2022-08-03T19:29:49Z"),
					},
				},
				"nine-mock": {
					{
						AssetURN:     "nine-mock",
						Status:       "CANCELLED",
						StatusReason: "User-friendly tertiary matrix",
						Timestamp:    newTS("2022-08-14T14:20:20Z"),
					},
				},
				"ten-mock": {
					{
						AssetURN:     "ten-mock",
						Status:       "CANCELLED",
						StatusReason: "Synergized bottom-line forecast",
						Timestamp:    newTS("2022-09-11T07:40:11Z"),
					},
				},
				"twelfth-mock": {
					{
						AssetURN:     "twelfth-mock",
						Status:       "TERMINATED",
						StatusReason: "Enterprise-wide interactive Graphical User Interface",
						Timestamp:    newTS("2022-04-15T00:02:25Z"),
					},
				},
			},
		},
		{
			name: "AssetURNs=c-demo-kafka;NewerThan=2022-03-08",
			flt:  asset.ProbesFilter{AssetURNs: []string{"c-demo-kafka"}, NewerThan: newTS("2022-03-08T00:00:00Z")},
			expected: map[string][]asset.Probe{
				"c-demo-kafka": {
					{
						AssetURN:  "c-demo-kafka",
						Status:    "SUCCESS",
						Timestamp: newTS("2022-03-08T09:58:43Z"),
					},
				},
			},
		},
		{
			name: "AssetURNs=c-demo-kafka;MaxRows=1",
			flt:  asset.ProbesFilter{AssetURNs: []string{"c-demo-kafka"}, MaxRows: 1},
			expected: map[string][]asset.Probe{
				"c-demo-kafka": {
					{
						AssetURN:  "c-demo-kafka",
						Status:    "SUCCESS",
						Timestamp: newTS("2022-03-08T09:58:43Z"),
					},
				},
			},
		},
		{
			name: "AssetURNs=c-demo-kafka,e-test-grant2;MaxRows=1",
			flt:  asset.ProbesFilter{AssetURNs: []string{"c-demo-kafka", "e-test-grant2"}, MaxRows: 1},
			expected: map[string][]asset.Probe{
				"c-demo-kafka": {
					{
						AssetURN:  "c-demo-kafka",
						Status:    "SUCCESS",
						Timestamp: newTS("2022-03-08T09:58:43Z"),
					},
				},
				"e-test-grant2": {
					{
						AssetURN:     "e-test-grant2",
						Status:       "TERMINATED",
						StatusReason: "Digitized asynchronous knowledge user",
						Timestamp:    newTS("2022-09-08T12:16:42Z"),
					},
				},
			},
		},
		{
			name: "NewerThan=2022-08-14;MaxRows=1",
			flt:  asset.ProbesFilter{NewerThan: newTS("2022-08-14T00:00:00Z"), MaxRows: 1},
			expected: map[string][]asset.Probe{
				"f-john-test-001": {
					{
						AssetURN:  "f-john-test-001",
						Status:    "CANCELLED",
						Timestamp: newTS("2022-09-23T14:39:57Z"),
					},
				},
				"ten-mock": {
					{
						AssetURN:     "ten-mock",
						Status:       "CANCELLED",
						StatusReason: "Synergized bottom-line forecast",
						Timestamp:    newTS("2022-09-11T07:40:11Z"),
					},
				},
				"e-test-grant2": {
					{
						AssetURN:     "e-test-grant2",
						Status:       "TERMINATED",
						StatusReason: "Digitized asynchronous knowledge user",
						Timestamp:    newTS("2022-09-08T12:16:42Z"),
					},
				},
				"nine-mock": {
					{
						AssetURN:     "nine-mock",
						Status:       "CANCELLED",
						StatusReason: "User-friendly tertiary matrix",
						Timestamp:    newTS("2022-08-14T14:20:20Z"),
					},
				},
				"h-test-new-kafka": {
					{
						AssetURN:     "h-test-new-kafka",
						Status:       "SUCCESS",
						StatusReason: "User-friendly systematic neural-net",
						Timestamp:    newTS("2022-08-14T03:04:44Z"),
					},
				},
			},
		},
		{
			name: "OlderThan=2021-11-08;MaxRows=1",
			flt:  asset.ProbesFilter{OlderThan: newTS("2021-11-08T00:00:00Z"), MaxRows: 1},
			expected: map[string][]asset.Probe{
				"i-undefined-dfgdgd-avi": {
					{
						AssetURN:     "i-undefined-dfgdgd-avi",
						Status:       "SUCCESS",
						StatusReason: "Persevering composite workforce",
						Timestamp:    newTS("2021-11-07T12:16:41Z"),
					},
				},
				"e-test-grant2": {
					{
						AssetURN:     "e-test-grant2",
						Status:       "TERMINATED",
						StatusReason: "Digitized asynchronous knowledge user",
						Timestamp:    newTS("2022-09-08T12:16:42Z"),
					},
				},
			},
		},
		{
			name: "AssetURNs=c-demo-kafka,e-test-grant2,nine-mock;MaxRows=1;NewerThan=2022-08-14;OlderThan=2022-09-11",
			flt: asset.ProbesFilter{
				AssetURNs: []string{"c-demo-kafka", "e-test-grant2", "nine-mock"},
				NewerThan: newTS("2022-08-14T00:00:00Z"),
				OlderThan: newTS("2022-09-11T00:00:00Z"),
				MaxRows:   1,
			},
			expected: map[string][]asset.Probe{
				"e-test-grant2": {
					{
						AssetURN:     "e-test-grant2",
						Status:       "TERMINATED",
						StatusReason: "Digitized asynchronous knowledge user",
						Timestamp:    newTS("2022-09-08T12:16:42Z"),
					},
				},
				"nine-mock": {
					{
						AssetURN:     "nine-mock",
						Status:       "CANCELLED",
						StatusReason: "User-friendly tertiary matrix",
						Timestamp:    newTS("2022-08-14T14:20:20Z"),
					},
				},
			},
		},
	}
	for _, tc := range cases {
		r.Run(tc.name, func() {
			actual, err := r.repository.GetProbesWithFilter(r.ctx, tc.flt)
			r.NoError(err)

			r.ElementsMatch(keys(tc.expected), keys(actual), "Mismatch of URN keys in map")
			for urn, expPrbs := range tc.expected {
				actPrbs, ok := actual[urn]
				if !ok || r.Lenf(actPrbs, len(expPrbs), "Mismatch in length of assets for URN '%s'", urn) {
					continue
				}

				for i := range actPrbs {
					r.assertProbe(r.T(), expPrbs[i], actPrbs[i])
				}
			}
		})
	}
}

func (r *AssetRepositoryTestSuite) insertProbes(t *testing.T) {
	t.Helper()

	probesJSON, err := os.ReadFile("./testdata/mock-probes-data.json")
	r.Require().NoError(err)

	var probes []asset.Probe
	r.Require().NoError(json.Unmarshal(probesJSON, &probes))

	for _, p := range probes {
		r.Require().NoError(r.repository.AddProbe(r.ctx, p.AssetURN, &p))
	}
}

func (r *AssetRepositoryTestSuite) assertAsset(expectedAsset, actualAsset *asset.Asset) bool {
	if expectedAsset == nil && actualAsset == nil {
		return true
	}
	if expectedAsset == nil {
		return false
	}
	if actualAsset == nil {
		return false
	}

	// sanitize time to make the assets comparable
	expectedAsset.CreatedAt = time.Time{}
	expectedAsset.UpdatedAt = time.Time{}
	expectedAsset.RefreshedAt = &time.Time{}
	expectedAsset.UpdatedBy.CreatedAt = time.Time{}
	expectedAsset.UpdatedBy.UpdatedAt = time.Time{}

	actualAsset.CreatedAt = time.Time{}
	actualAsset.UpdatedAt = time.Time{}
	actualAsset.RefreshedAt = &time.Time{}
	actualAsset.UpdatedBy.CreatedAt = time.Time{}
	actualAsset.UpdatedBy.UpdatedAt = time.Time{}

	return r.Equal(expectedAsset, actualAsset)
}

func clearTimestamps(ast *asset.Asset) {
	ast.UpdatedBy.CreatedAt = time.Time{}
	ast.UpdatedBy.UpdatedAt = time.Time{}
	ast.CreatedAt = time.Time{}
	ast.UpdatedAt = time.Time{}
}

func (r *AssetRepositoryTestSuite) assertProbe(t *testing.T, expected, actual asset.Probe) bool {
	t.Helper()

	return r.Equal(expected.AssetURN, actual.AssetURN) &&
		r.Equal(expected.Status, actual.Status) &&
		r.Equal(expected.StatusReason, actual.StatusReason) &&
		r.Equal(expected.Metadata, actual.Metadata) &&
		r.Equal(expected.Timestamp, actual.Timestamp)
}

func TestAssetRepository(t *testing.T) {
	suite.Run(t, &AssetRepositoryTestSuite{})
}

func stripUserID(u user.User) user.User {
	u.ID = ""
	return u
}
