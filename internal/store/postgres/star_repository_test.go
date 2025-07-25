package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/goto/compass/core/star"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/testutils"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/suite"
)

type StarRepositoryTestSuite struct {
	suite.Suite
	ctx             context.Context
	client          *postgres.Client
	repository      *postgres.StarRepository
	userRepository  *postgres.UserRepository
	assetRepository *postgres.AssetRepository
}

func (r *StarRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewLogrus()
	r.client, err = newTestClient(r.T(), logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewStarRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
	r.userRepository, err = postgres.NewUserRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
	r.assetRepository, err = postgres.NewAssetRepository(r.client, r.userRepository, postgres.DEFAULT_MAX_RESULT_SIZE, "")
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *StarRepositoryTestSuite) TestCreate() {
	ownerEmail := "test-create@gotocompany.com"

	r.Run("return no error if successfully create star", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		userID, err := createUser(r.userRepository, "user@gotocompany.com")
		r.NoError(err)

		createdAsset, err := createAsset(r.assetRepository, userID, ownerEmail, "asset-urn-1", "table")
		r.NoError(err)

		id, err := r.repository.Create(r.ctx, userID, createdAsset.ID)
		r.NoError(err)
		r.NotEmpty(id)
	})

	r.Run("return ErrEmptyUserID if user id is empty", func() {
		id, err := r.repository.Create(r.ctx, "", "")
		r.ErrorIs(err, star.ErrEmptyUserID)
		r.Empty(id)
	})

	r.Run("return ErrEmptyAssetID error if asset id is empty", func() {
		id, err := r.repository.Create(r.ctx, "user-id", "")
		r.ErrorIs(err, star.ErrEmptyAssetID)
		r.Empty(id)
	})

	r.Run("return invalid error if user id is not uuid", func() {
		id, err := r.repository.Create(r.ctx, "user-id", "asset-id")
		r.ErrorIs(err, star.InvalidError{UserID: "user-id"})
		r.Empty(id)
	})

	r.Run("return invalid error if asset id is not uuid", func() {
		uid := uuid.NewString()
		id, err := r.repository.Create(r.ctx, uid, "asset-id")
		r.ErrorIs(err, star.InvalidError{AssetID: "asset-id"})
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if starred asset is already exist", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		userID, err := createUser(r.userRepository, "user@gotocompany.com")
		r.NoError(err)

		createdAsset, err := createAsset(r.assetRepository, userID, ownerEmail, "asset-urn-1", "table")
		r.NoError(err)

		id, err := r.repository.Create(r.ctx, userID, createdAsset.ID)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.Create(r.ctx, userID, createdAsset.ID)
		r.ErrorIs(err, star.DuplicateRecordError{UserID: userID, AssetID: createdAsset.ID})
		r.Empty(id)
	})
}

func (r *StarRepositoryTestSuite) TestGetStargazers() {
	ownerEmail := "test-getstargazers@gotocompany.com"

	defaultCfg := star.Filter{}
	r.Run("return ErrEmptyAssetID if asset id is empty", func() {
		users, err := r.repository.GetStargazers(r.ctx, defaultCfg, "")
		r.ErrorIs(err, star.ErrEmptyAssetID)
		r.Empty(users)
	})

	r.Run("return invalid error if asset id is not uuid", func() {
		users, err := r.repository.GetStargazers(r.ctx, defaultCfg, "asset-id")
		r.ErrorIs(err, star.InvalidError{AssetID: "asset-id"})
		r.Empty(users)
	})

	r.Run("return not found error if starred asset not found in db", func() {
		assetID := uuid.NewString()
		users, err := r.repository.GetStargazers(r.ctx, defaultCfg, assetID)
		r.ErrorIs(err, star.NotFoundError{AssetID: assetID})
		r.Empty(users)
	})

	r.Run("return list of users that star an asset if get users success", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)
		userID1, err := createUser(r.userRepository, "user@gotocompany.com")
		r.NoError(err)

		createdAsset1, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-1", "table")
		r.NoError(err)

		id, err := r.repository.Create(r.ctx, userID1, createdAsset1.ID)
		r.NoError(err)
		r.NotEmpty(id)

		userID2, err := createUser(r.userRepository, "admin@gotocompany.com")
		r.NoError(err)

		id, err = r.repository.Create(r.ctx, userID2, createdAsset1.ID)
		r.NoError(err)
		r.NotEmpty(id)

		actualUsers, err := r.repository.GetStargazers(r.ctx, defaultCfg, createdAsset1.ID)
		r.NoError(err)

		actualUserIDs := []string{}
		for _, user := range actualUsers {
			actualUserIDs = append(actualUserIDs, user.ID)
		}

		r.Len(actualUserIDs, 2)
		r.Contains(actualUserIDs, userID1)
		r.Contains(actualUserIDs, userID2)
	})

	r.Run("return limited paginated list of users that star an asset if get users success", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		var assetID string

		for i := 1; i < 20; i++ {
			userEmail := fmt.Sprintf("user%d@gotocompany.com", i)
			userID, err := createUser(r.userRepository, userEmail)
			r.NoError(err)
			createdAsset, err := createAsset(r.assetRepository, userID, ownerEmail, "asset-urn-1", "table")
			r.NoError(err)

			id, err := r.repository.Create(r.ctx, userID, createdAsset.ID)
			r.NoError(err)
			r.NotEmpty(id)

			assetID = createdAsset.ID
		}

		cfg := star.Filter{Size: 7}
		actualUsers, err := r.repository.GetStargazers(r.ctx, cfg, assetID)
		r.NoError(err)

		r.Len(actualUsers, 7)
	})
}

func (r *StarRepositoryTestSuite) TestGetAllAssetsByUserID() {
	ownerEmail := "test-getallbyuserid@gotocompany.com"

	defaultCfg := star.Filter{}
	r.Run("return invalid error if user id is empty", func() {
		assets, err := r.repository.GetAllAssetsByUserID(r.ctx, defaultCfg, "")
		r.ErrorIs(err, star.ErrEmptyUserID)
		r.Empty(assets)
	})

	r.Run("return invalid error if user id is not uuid", func() {
		assets, err := r.repository.GetAllAssetsByUserID(r.ctx, defaultCfg, "user-id")
		r.ErrorIs(err, star.InvalidError{UserID: "user-id"})
		r.Empty(assets)
	})

	r.Run("return not found error if starred asset not found in db", func() {
		randomUserID := uuid.NewString()
		assets, err := r.repository.GetAllAssetsByUserID(r.ctx, defaultCfg, randomUserID)
		r.ErrorIs(err, star.NotFoundError{UserID: randomUserID})
		r.Empty(assets)
	})

	r.Run("return list of starred assets if get by user id success", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		userID1, err := createUser(r.userRepository, "user@gotocompany.com")
		r.NoError(err)

		createdAsset1, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-1", "table")
		r.NoError(err)
		id, err := r.repository.Create(r.ctx, userID1, createdAsset1.ID)
		r.NoError(err)
		r.NotEmpty(id)

		createdAsset2, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-2", "table")
		r.NoError(err)
		id, err = r.repository.Create(r.ctx, userID1, createdAsset2.ID)
		r.NoError(err)
		r.NotEmpty(id)

		createdAsset3, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-3", "table")
		r.NoError(err)
		id, err = r.repository.Create(r.ctx, userID1, createdAsset3.ID)
		r.NoError(err)
		r.NotEmpty(id)

		actualAssets, err := r.repository.GetAllAssetsByUserID(r.ctx, defaultCfg, userID1)
		r.NoError(err)

		assetIDs := []string{}
		for _, asset := range actualAssets {
			assetIDs = append(assetIDs, asset.ID)
		}

		r.Len(actualAssets, 3)
		r.Contains(assetIDs, createdAsset1.ID)
		r.Contains(assetIDs, createdAsset2.ID)
		r.Contains(assetIDs, createdAsset3.ID)
	})

	r.Run("return limited paginated list of starred assets if get by user id success", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		userID, err := createUser(r.userRepository, "user@gotocompany.com")
		r.NoError(err)

		for i := 1; i < 20; i++ {
			starURN := fmt.Sprintf("asset-urn-%d", i)
			createdAsset, err := createAsset(r.assetRepository, userID, ownerEmail, starURN, "table")
			r.NoError(err)
			id, err := r.repository.Create(r.ctx, userID, createdAsset.ID)
			r.NoError(err)
			r.NotEmpty(id)
		}

		cfg := star.Filter{Size: 7}
		actualAssets, err := r.repository.GetAllAssetsByUserID(r.ctx, cfg, userID)
		r.NoError(err)
		r.NoError(err)

		r.Len(actualAssets, 7)
	})
}

func (r *StarRepositoryTestSuite) TestGetAssetByUserID() {
	ownerEmail := "test-getbyuserid@gotocompany.com"

	r.Run("return ErrEmptyUserID if user id is empty", func() {
		ast, err := r.repository.GetAssetByUserID(r.ctx, "", "")
		r.ErrorIs(err, star.ErrEmptyUserID)
		r.Empty(ast)
	})

	r.Run("return ErrEmptyAssetID if asset id is empty", func() {
		ast, err := r.repository.GetAssetByUserID(r.ctx, "user-id", "")
		r.ErrorIs(err, star.ErrEmptyAssetID)
		r.Empty(ast)
	})

	r.Run("return invalid error if user id is not uuid", func() {
		ast, err := r.repository.GetAssetByUserID(r.ctx, "user-id", "asset-id")
		r.ErrorIs(err, star.InvalidError{UserID: "user-id"})
		r.Empty(ast)
	})

	r.Run("return invalid error if asset id is not uuid", func() {
		uid := uuid.NewString()
		ast, err := r.repository.GetAssetByUserID(r.ctx, uid, "asset-id")
		r.ErrorIs(err, star.InvalidError{AssetID: "asset-id"})
		r.Empty(ast)
	})

	r.Run("return not found error if starred asset not found in db", func() {
		randomUserID := uuid.NewString()
		assetID := uuid.NewString()
		ast, err := r.repository.GetAssetByUserID(r.ctx, randomUserID, assetID)
		r.ErrorIs(err, star.NotFoundError{AssetID: assetID, UserID: randomUserID})
		r.Empty(ast)
	})

	r.Run("return the starred assets if get user starred asset success", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		userID1, err := createUser(r.userRepository, "user@gotocompany.com")
		r.NoError(err)

		createdAsset1, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-1", "table")
		r.NoError(err)
		id, err := r.repository.Create(r.ctx, userID1, createdAsset1.ID)
		r.NoError(err)
		r.NotEmpty(id)

		createdAsset2, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-2", "table")
		r.NoError(err)
		id, err = r.repository.Create(r.ctx, userID1, createdAsset2.ID)
		r.NoError(err)
		r.NotEmpty(id)

		createdAsset3, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-3", "table")
		r.NoError(err)
		id, err = r.repository.Create(r.ctx, userID1, createdAsset3.ID)
		r.NoError(err)
		r.NotEmpty(id)

		actualAsset, err := r.repository.GetAssetByUserID(r.ctx, userID1, createdAsset2.ID)
		r.NoError(err)

		r.Equal(createdAsset2.ID, actualAsset.ID)
	})
}

func (r *StarRepositoryTestSuite) TestDelete() {
	ownerEmail := "test-delete@gotocompany.com"

	r.Run("return invalid error if user id is empty", func() {
		err := r.repository.Delete(r.ctx, "", "")
		r.ErrorIs(err, star.ErrEmptyUserID)
	})

	r.Run("return invalid error if asset id is empty", func() {
		err := r.repository.Delete(r.ctx, "user-id", "")
		r.ErrorIs(err, star.ErrEmptyAssetID)
	})

	r.Run("return invalid error if user id is not uuid", func() {
		err := r.repository.Delete(r.ctx, "user-id", "asset-id")
		r.ErrorIs(err, star.InvalidError{UserID: "user-id"})
	})

	r.Run("return invalid error if asset id is not uuid", func() {
		uid := uuid.NewString()
		err := r.repository.Delete(r.ctx, uid, "asset-id")
		r.ErrorIs(err, star.InvalidError{AssetID: "asset-id"})
	})

	r.Run("return not found error if starred asset not found in db", func() {
		randomUserID := uuid.NewString()
		assetID := uuid.NewString()
		err := r.repository.Delete(r.ctx, randomUserID, assetID)
		r.ErrorIs(err, star.NotFoundError{AssetID: assetID, UserID: randomUserID})
	})

	r.Run("return nil if successfully unstar an asset", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		userID1, err := createUser(r.userRepository, "user@gotocompany.com")
		r.NoError(err)

		createdAsset1, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-1", "table")
		r.NoError(err)
		id, err := r.repository.Create(r.ctx, userID1, createdAsset1.ID)
		r.NoError(err)
		r.NotEmpty(id)

		createdAsset2, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-2", "table")
		r.NoError(err)
		id, err = r.repository.Create(r.ctx, userID1, createdAsset2.ID)
		r.NoError(err)
		r.NotEmpty(id)

		createdAsset3, err := createAsset(r.assetRepository, userID1, ownerEmail, "asset-urn-3", "table")
		r.NoError(err)
		id, err = r.repository.Create(r.ctx, userID1, createdAsset3.ID)
		r.NoError(err)
		r.NotEmpty(id)

		actualAssets, err := r.repository.GetAllAssetsByUserID(r.ctx, star.Filter{}, userID1)
		r.NoError(err)

		r.Len(actualAssets, 3)

		err = r.repository.Delete(r.ctx, userID1, createdAsset3.ID)
		r.NoError(err)

		actualAssets, err = r.repository.GetAllAssetsByUserID(r.ctx, star.Filter{}, userID1)
		r.NoError(err)

		assetIDs := []string{}
		for _, asset := range actualAssets {
			assetIDs = append(assetIDs, asset.ID)
		}

		r.Len(actualAssets, 2)
		r.Contains(assetIDs, createdAsset1.ID)
		r.Contains(assetIDs, createdAsset2.ID)
	})
}

func TestStarRepository(t *testing.T) {
	suite.Run(t, &StarRepositoryTestSuite{})
}
