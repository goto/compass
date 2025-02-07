package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
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
	"github.com/stretchr/testify/suite"
)

var defaultAssetUpdaterUserID = uuid.NewString()

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
	users := []user.User{}

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

func (r *AssetRepositoryTestSuite) BeforeTest(suiteName, testName string) {
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

		id, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		r.Require().NotEmpty(id)
		ast.ID = id
		assets = append(assets, ast)
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
	assets := r.insertRecord()

	r.Run("should return error if SortBy key is invalid", func() {
		_, err := r.repository.GetAll(r.ctx, asset.Filter{
			SortBy: "wrong key name",
		})
		r.Require().NotNil(err)
		r.Require().ErrorContains(err, "error getting asset list")
	})

	r.Run("should return all assets without filtering based on size", func() {
		expectedSize := 12

		results, err := r.repository.GetAll(r.ctx, asset.Filter{})
		r.Require().NoError(err)
		r.Require().Len(results, expectedSize)
		for i := 0; i < expectedSize; i++ {
			r.assertAsset(&assets[i], &results[i])
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
			r.assertAsset(&assets[i], &results[i])
		}
	})

	r.Run("should fetch assets by offset defined in GetConfig.Offset", func() {
		offset := 2

		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Offset: offset,
		})
		r.Require().NoError(err)
		for i := offset; i > len(results)+offset; i++ {
			r.assertAsset(&assets[i], &results[i-offset])
		}
	})

	r.Run("should filter using type", func() {
		results, err := r.repository.GetAll(r.ctx, asset.Filter{
			Types:         []asset.Type{asset.Type("table")},
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
}

func (r *AssetRepositoryTestSuite) TestGetTypes() {
	_ = r.insertRecord()

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
				typeTopic:     3,
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
			Type:      typ,
			Service:   service[0],
			UpdatedBy: r.users[0],
		}
		id, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		r.Require().NotEmpty(id)
		ast.ID = id
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

func (r *AssetRepositoryTestSuite) TestGetByID() {
	r.Run("return error from client if asset not an uuid", func() {
		_, err := r.repository.GetByID(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
	})

	r.Run("return NotFoundError if asset does not exist", func() {
		uuid := "2aabb450-f986-44e2-a6db-7996861d5004"
		_, err := r.repository.GetByID(r.ctx, uuid)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuid})
	})

	r.Run("return correct asset from db", func() {
		asset1 := asset.Asset{
			URN:       "urn-gbi-1",
			Type:      "table",
			Service:   "bigquery",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
		}
		asset2 := asset.Asset{
			URN:       "urn-gbi-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
		}

		var err error
		id, err := r.repository.Upsert(r.ctx, &asset1)
		r.Require().NoError(err)
		r.NotEmpty(id)
		asset1.ID = id

		id, err = r.repository.Upsert(r.ctx, &asset2)
		r.Require().NoError(err)
		r.NotEmpty(id)
		asset2.ID = id

		result, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		asset2.UpdatedBy = r.users[1]
		r.assertAsset(&asset2, &result)
	})

	r.Run("return owners if any", func() {
		ast := asset.Asset{
			URN:     "urn-gbi-3",
			Type:    "table",
			Service: "bigquery",
			Owners: []user.User{
				r.users[1],
				r.users[2],
			},
			UpdatedBy: r.users[1],
		}

		id, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		r.Require().NotEmpty(id)
		ast.ID = id

		result, err := r.repository.GetByID(r.ctx, ast.ID)
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
			Type:      "table",
			Service:   "bigquery",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
		}
		asset2 := asset.Asset{
			URN:       "urn-gbi-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: r.users[1],
		}

		id, err := r.repository.Upsert(r.ctx, &asset1)
		r.Require().NoError(err)
		r.NotEmpty(id)
		asset1.ID = id

		id, err = r.repository.Upsert(r.ctx, &asset2)
		r.Require().NoError(err)
		r.NotEmpty(id)
		asset2.ID = id

		result, err := r.repository.GetByURN(r.ctx, "urn-gbi-2")
		r.NoError(err)
		r.assertAsset(&asset2, &result)
	})

	r.Run("return owners if any", func() {
		ast := asset.Asset{
			URN:     "urn-gbi-3",
			Type:    "table",
			Service: "bigquery",
			Owners: []user.User{
				r.users[1],
				r.users[2],
			},
			UpdatedBy: r.users[1],
		}

		_, err := r.repository.Upsert(r.ctx, &ast)
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
	// v0.1
	astVersioning := asset.Asset{
		URN:         assetURN,
		Type:        "table",
		Service:     "bigquery",
		UpdatedBy:   r.users[1],
		RefreshedAt: &currentTime,
	}

	id, err := r.repository.Upsert(r.ctx, &astVersioning)
	r.Require().NoError(err)
	r.Require().NotEmpty(id)
	astVersioning.ID = id

	// v0.2
	astVersioning.Description = "new description in v0.2"
	id, err = r.repository.Upsert(r.ctx, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

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
	id, err = r.repository.Upsert(r.ctx, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

	// v0.4
	astVersioning.Data = map[string]interface{}{
		"data1": float64(12345),
	}
	id, err = r.repository.Upsert(r.ctx, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

	// v0.5
	astVersioning.Labels = map[string]string{
		"key1": "value1",
	}

	id, err = r.repository.Upsert(r.ctx, &astVersioning)
	r.Require().NoError(err)
	r.Require().Equal(id, astVersioning.ID)

	r.Run("should return current version of an assets", func() {
		expected := asset.Asset{
			ID:          astVersioning.ID,
			URN:         assetURN,
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
			Type:        "table",
			Service:     "bigquery",
			Description: "new description in v0.2",
			Version:     "0.3",
			Changelog: diff.Changelog{
				diff.Change{Type: "create", Path: []string{"owners", "0", "email"}, From: interface{}(nil), To: "user@gotocompany.com"},
				diff.Change{Type: "create", Path: []string{"owners", "1", "email"}, From: interface{}(nil), To: "meteor@gotocompany.com"},
			},
			UpdatedBy: r.users[1],
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
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: r.users[1],
		}
		id, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		r.Require().NotEmpty(id)
		ast.ID = id

		for i := 2; i < 100; i++ {
			ast.Description = "new description in v0." + strconv.Itoa(i)
			id, err = r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.Require().Equal(id, ast.ID)
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

		assetVersions, err := r.repository.GetVersionHistory(r.ctx, asset.Filter{Size: 3, Offset: 86}, ast.ID)
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
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: r.users[1],
		}
		id, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		r.Require().NotEmpty(id)
		ast.ID = id

		for i := 2; i < 100; i++ {
			ast.Description = "new description in v0." + strconv.Itoa(i)
			id, err = r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.Require().Equal(id, ast.ID)
		}

		assetVersions, err := r.repository.GetVersionHistory(r.ctx, asset.Filter{Size: 0, Offset: 86}, ast.ID)
		r.NoError(err)
		r.Equal(defaultGetMaxSize, len(assetVersions))
	})

	r.Run("should return error if invalid uuid is passed", func() {
		assetURN := "invalid uuid"
		_, err := r.repository.GetVersionHistory(r.ctx, asset.Filter{Size: 3, Offset: 86}, assetURN)
		r.NotNil(err)
		r.Equal(asset.InvalidError{AssetID: assetURN}, err)
	})
}

func (r *AssetRepositoryTestSuite) TestUpsert() {
	refreshedAtTime := time.Date(2024, time.August, 20, 8, 19, 49, 0, time.UTC)
	r.Run("on insert", func() {
		r.Run("set ID to asset and version to base version", func() {
			ast := asset.Asset{
				URN:         "urn-u-1",
				Type:        "table",
				Service:     "bigquery",
				URL:         "https://sample-url.com",
				UpdatedBy:   r.users[0],
				RefreshedAt: &refreshedAtTime,
			}
			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Equal(asset.BaseVersion, ast.Version)
			r.NoError(err)
			r.NotEmpty(id)
			r.NotEmpty(ast.CreatedAt)
			r.NotEmpty(ast.UpdatedAt)
			ast.ID = id

			assetInDB, err := r.repository.GetByID(r.ctx, ast.ID)
			r.Require().NoError(err)
			r.NotEqual(time.Time{}, assetInDB.CreatedAt)
			r.NotEqual(time.Time{}, assetInDB.UpdatedAt)
			r.assertAsset(&ast, &assetInDB)

			ast2 := ast
			ast2.RefreshedAt = &refreshedAtTime
			ast2.Description = "create a new version" // to force fetch from asset_versions.
			_, err = r.repository.Upsert(r.ctx, &ast2)
			r.NoError(err)
			r.Greater(ast2.UpdatedAt.UnixNano(), ast.UpdatedAt.UnixNano())
			assetv1, err := r.repository.GetByVersionWithID(r.ctx, ast.ID, asset.BaseVersion)
			r.NoError(err)
			r.Equal("0.1", assetv1.Version)
		})

		r.Run("should store owners if any", func() {
			ast := asset.Asset{
				URN:     "urn-u-3",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					r.users[1],
					{Email: r.users[2].Email},
					{ID: r.users[1].ID}, // should get deduplicated by ID
				},
				UpdatedBy: r.users[0],
			}

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.Require().NotEmpty(id)
			ast.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)

			r.Len(actual.Owners, 2)
			r.Equal(r.users[1].ID, actual.Owners[0].ID)
			r.Equal(r.users[2].ID, actual.Owners[1].ID)
		})

		r.Run("should create owners as users if they do not exist yet", func() {
			ast := asset.Asset{
				URN:     "urn-u-3a",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					{Email: "newuser@example.com"},
					{Email: "newuser2@example.com"},
					{Email: "newuser@example.com"}, // should get deduplicated by ID on fetch user by email
				},
				UpdatedBy: r.users[0],
			}

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.NotEmpty(id)

			actual, err := r.repository.GetByID(r.ctx, id)
			r.NoError(err)

			r.Len(actual.Owners, 2)
			r.Equal(ast.Owners[0].Email, actual.Owners[0].Email)
		})
	})

	r.Run("on update", func() {
		r.Run("should not create nor updating the asset if asset is identical", func() {
			ast := asset.Asset{
				URN:         "urn-u-2",
				Type:        "table",
				Service:     "bigquery",
				UpdatedBy:   r.users[0],
				RefreshedAt: &refreshedAtTime,
				Version:     "0.1",
			}
			identicalAsset := ast

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.NotEmpty(id)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, &identicalAsset)
			r.Require().NoError(err)
			r.NotEmpty(id)
			identicalAsset.ID = id

			r.Equal(ast.ID, identicalAsset.ID)
			r.Equal(ast.Version, identicalAsset.Version)
		})

		r.Run("should same asset version if asset only has different at RefreshedAt", func() {
			oneDayAgoRefreshedAtTime := refreshedAtTime.AddDate(0, 0, -1)
			ast := asset.Asset{
				URN:         "urn-u-2",
				Type:        "table",
				Service:     "bigquery",
				URL:         "https://sample-url-old.com",
				UpdatedBy:   r.users[0],
				RefreshedAt: &oneDayAgoRefreshedAtTime,
				Version:     "0.1",
			}

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.NotEmpty(id)
			ast.ID = id

			updated := ast
			updated.RefreshedAt = &refreshedAtTime

			id, err = r.repository.Upsert(r.ctx, &updated)
			r.Require().NoError(err)
			r.NotEmpty(id)
			updated.ID = id

			r.Equal(ast.ID, updated.ID)

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)

			r.Equal(updated.RefreshedAt, actual.RefreshedAt)
			r.Equal(ast.Version, actual.Version)
		})

		r.Run("should update the asset version if asset is not identical", func() {
			ast := asset.Asset{
				URN:       "urn-u-2",
				Type:      "table",
				Service:   "bigquery",
				URL:       "https://sample-url-old.com",
				UpdatedBy: r.users[0],
				Version:   "0.1",
			}

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.NotEmpty(id)
			ast.ID = id

			updated := ast
			updated.URL = "https://sample-url.com"

			id, err = r.repository.Upsert(r.ctx, &updated)
			r.Require().NoError(err)
			r.NotEmpty(id)
			updated.ID = id

			r.Equal(ast.ID, updated.ID)

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)

			r.Equal(updated.URL, actual.URL)
			r.NotEqual(ast.Version, actual.Version)
		})

		r.Run("should delete old owners if it does not exist on new asset", func() {
			ast := asset.Asset{
				URN:     "urn-u-4",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
					stripUserID(r.users[2]),
				},
				UpdatedBy: r.users[0],
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				stripUserID(r.users[2]),
			}

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.NotEmpty(id)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, &newAsset)
			r.Require().NoError(err)
			r.NotEmpty(id)
			newAsset.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)
			r.Len(actual.Owners, len(newAsset.Owners))
			r.Equal(r.users[2].ID, actual.Owners[0].ID)
		})

		r.Run("should create new owners if it does not exist on old asset", func() {
			ast := asset.Asset{
				URN:     "urn-u-4",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				UpdatedBy: r.users[0],
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				stripUserID(r.users[1]),
				stripUserID(r.users[2]),
			}

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.NotEmpty(id)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, &newAsset)
			r.Require().NoError(err)
			r.NotEmpty(id)
			newAsset.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)
			r.Len(actual.Owners, len(newAsset.Owners))
			r.Equal(r.users[1].ID, actual.Owners[0].ID)
			r.Equal(r.users[2].ID, actual.Owners[1].ID)
		})

		r.Run("should create users from owners if owner emails do not exist yet", func() {
			ast := asset.Asset{
				URN:     "urn-u-4a",
				Type:    "table",
				Service: "bigquery",
				Owners: []user.User{
					stripUserID(r.users[1]),
				},
				UpdatedBy: r.users[0],
			}
			newAsset := ast
			newAsset.Owners = []user.User{
				stripUserID(r.users[1]),
				{Email: "newuser@example.com"},
			}

			id, err := r.repository.Upsert(r.ctx, &ast)
			r.Require().NoError(err)
			r.NotEmpty(id)
			ast.ID = id

			id, err = r.repository.Upsert(r.ctx, &newAsset)
			r.Require().NoError(err)
			r.NotEmpty(id)
			newAsset.ID = id

			actual, err := r.repository.GetByID(r.ctx, ast.ID)
			r.NoError(err)
			r.Len(actual.Owners, len(newAsset.Owners))
			r.NotEmpty(actual.Owners[0].ID)
			r.Equal(r.users[1].ID, actual.Owners[0].ID)
			r.NotEmpty(actual.Owners[1].ID)
			r.Equal(newAsset.Owners[1].Email, actual.Owners[1].Email)
		})
	})
}

func (r *AssetRepositoryTestSuite) TestUpsertRaceCondition() {
	r.Run("TestUpsertRaceCondition", func() {
		ast := asset.Asset{
			URN:       "urn-u-3",
			Type:      "table",
			Service:   "bigquery",
			URL:       "https://sample-url-old.com",
			UpdatedBy: r.users[0],
			Version:   "0.1",
		}

		id, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		r.NotEmpty(id)
		ast.ID = id

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
				_, err := r.repository.Upsert(r.ctx, &localAst)

				mu.Lock()
				results = append(results, err)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Check for errors
		for i, err := range results {
			fmt.Println("err", i, ": ", err)
			assert.NoError(r.T(), err, "Upsert should not fail under race conditions")
		}
	})
}

func (r *AssetRepositoryTestSuite) TestDeleteByID() {
	r.Run("return error from client if any", func() {
		err := r.repository.DeleteByID(r.ctx, "invalid-uuid")
		r.Error(err)
		r.Contains(err.Error(), "invalid asset id: \"invalid-uuid\"")
	})

	r.Run("return NotFoundError if asset does not exist", func() {
		uuid := "2aabb450-f986-44e2-a6db-7996861d5004"
		err := r.repository.DeleteByID(r.ctx, uuid)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: uuid})
	})

	r.Run("should delete correct asset", func() {
		asset1 := asset.Asset{
			URN:       "urn-del-1",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
		}
		asset2 := asset.Asset{
			URN:       "urn-del-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
		}

		var err error
		id, err := r.repository.Upsert(r.ctx, &asset1)
		r.Require().NoError(err)
		r.Require().NotEmpty(id)
		asset1.ID = id

		id, err = r.repository.Upsert(r.ctx, &asset2)
		r.Require().NoError(err)
		r.Require().NotEmpty(id)
		asset2.ID = id

		err = r.repository.DeleteByID(r.ctx, asset1.ID)
		r.NoError(err)

		_, err = r.repository.GetByID(r.ctx, asset1.ID)
		r.ErrorAs(err, &asset.NotFoundError{AssetID: asset1.ID})

		asset2FromDB, err := r.repository.GetByID(r.ctx, asset2.ID)
		r.NoError(err)
		r.Equal(asset2.ID, asset2FromDB.ID)

		// cleanup
		err = r.repository.DeleteByID(r.ctx, asset2.ID)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestDeleteByURN() {
	r.Run("return NotFoundError if asset does not exist", func() {
		urn := "urn-test-1"
		err := r.repository.DeleteByURN(r.ctx, urn)
		r.ErrorAs(err, &asset.NotFoundError{URN: urn})
	})

	r.Run("should delete correct asset", func() {
		asset1 := asset.Asset{
			URN:       "urn-del-1",
			Type:      "table",
			Service:   "bigquery",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
		}
		asset2 := asset.Asset{
			URN:       "urn-del-2",
			Type:      "topic",
			Service:   "kafka",
			Version:   asset.BaseVersion,
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
		}

		_, err := r.repository.Upsert(r.ctx, &asset1)
		r.Require().NoError(err)

		id, err := r.repository.Upsert(r.ctx, &asset2)
		r.Require().NoError(err)

		err = r.repository.DeleteByURN(r.ctx, asset1.URN)
		r.NoError(err)

		_, err = r.repository.GetByURN(r.ctx, asset1.URN)
		r.ErrorAs(err, &asset.NotFoundError{URN: asset1.URN})

		asset2FromDB, err := r.repository.GetByURN(r.ctx, asset2.URN)
		r.NoError(err)
		r.Equal(id, asset2FromDB.ID)

		// cleanup
		err = r.repository.DeleteByURN(r.ctx, asset2.URN)
		r.NoError(err)
	})
}

func (r *AssetRepositoryTestSuite) TestDeleteByQueryExpr() {
	refreshedAtTime := time.Date(2024, time.August, 20, 8, 19, 49, 0, time.UTC)
	r.Run("should delete correct asset", func() {
		oneYearAgoRefreshedAtTime := refreshedAtTime.AddDate(-1, 0, 0)
		asset1 := asset.Asset{
			URN:         "urn-del-1",
			Type:        "table",
			Service:     "bigquery",
			UpdatedBy:   user.User{ID: defaultAssetUpdaterUserID},
			RefreshedAt: &oneYearAgoRefreshedAtTime,
		}
		asset2 := asset.Asset{
			URN:         "urn-del-2",
			Type:        "topic",
			Service:     "kafka",
			Version:     asset.BaseVersion,
			UpdatedBy:   user.User{ID: defaultAssetUpdaterUserID},
			RefreshedAt: &oneYearAgoRefreshedAtTime,
		}

		_, err := r.repository.Upsert(r.ctx, &asset1)
		r.Require().NoError(err)

		id, err := r.repository.Upsert(r.ctx, &asset2)
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
		r.Equal(id, asset2FromDB.ID)

		// cleanup
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
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
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

		_, err := r.repository.Upsert(r.ctx, &ast)
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
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
		}
		probe := asset.Probe{
			Status:       "COMPLETED",
			StatusReason: "Sample Reason",
			Timestamp:    time.Now().Add(2 * time.Minute),
			Metadata: map[string]interface{}{
				"foo": "bar",
			},
		}

		_, err := r.repository.Upsert(r.ctx, &ast)
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
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
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

		_, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)

		err = r.repository.AddProbe(r.ctx, ast.URN, &probe)
		r.NoError(err)

		r.Equal(probeID, probe.ID)
	})

	r.Run("should populate Timestamp if empty", func() {
		ast := asset.Asset{
			URN:       "urn-add-probe-2",
			Type:      typeJob,
			Service:   "optimus",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
		}
		otherAst := asset.Asset{
			URN:       "urn-add-probe-3",
			Type:      typeJob,
			Service:   "airflow",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
		}
		probe := asset.Probe{
			Status: "RUNNING",
		}
		otherProbe := asset.Probe{
			Status: "RUNNING",
		}

		_, err := r.repository.Upsert(r.ctx, &ast)
		r.Require().NoError(err)
		_, err = r.repository.Upsert(r.ctx, &otherAst)
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
			Type:      asset.Type("job"),
			Service:   "airflow",
			UpdatedBy: user.User{ID: defaultAssetUpdaterUserID},
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

		_, err := r.repository.Upsert(r.ctx, &ast)
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
