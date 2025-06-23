package asset_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/asset/mocks"
	"github.com/goto/compass/internal/workermanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_GetAllAssets(t *testing.T) {
	type testCase struct {
		Description string
		Filter      asset.Filter
		WithTotal   bool
		Err         error
		ResultLen   int
		TotalCnt    uint32
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	testCases := []testCase{
		{
			Description: `should return error if asset repository get all return error and with total false`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{}, errors.New("unknown error"))
			},
			Err:       errors.New("unknown error"),
			ResultLen: 0,
			TotalCnt:  0,
		},
		{
			Description: `should return assets if asset repository get all return no error and with total false`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{
					{
						ID: "some-id",
					},
				}, nil)
			},
			Err:       errors.New("unknown error"),
			ResultLen: 1,
			TotalCnt:  0,
		},
		{
			Description: `should return error if asset repository get count return error and with total true`,
			WithTotal:   true,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{
					{
						ID: "some-id",
					},
				}, nil)
				ar.EXPECT().GetCount(ctx, asset.Filter{}).Return(0, errors.New("unknown error"))
			},
			Err:       errors.New("unknown error"),
			ResultLen: 0,
			TotalCnt:  0,
		},
		{
			Description: `should return no error if asset repository get count return no error and with total true`,
			WithTotal:   true,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{
					{
						ID: "some-id",
					},
				}, nil)
				ar.EXPECT().GetCount(ctx, asset.Filter{}).Return(1, nil)
			},
			Err:       nil,
			ResultLen: 1,
			TotalCnt:  1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     mockAssetRepo,
				DiscoveryRepo: mockDiscoveryRepo,
				LineageRepo:   mockLineageRepo,
			})
			defer cancel()

			got, cnt, err := svc.GetAllAssets(ctx, tc.Filter, tc.WithTotal)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
			if tc.ResultLen != len(got) {
				t.Fatalf("got result len %v, expected result len was %v", len(got), tc.ResultLen)
			}
			if tc.TotalCnt != cnt {
				t.Fatalf("got total count %v, expected total count was %v", cnt, tc.TotalCnt)
			}
		})
	}
}

func TestService_GetTypes(t *testing.T) {
	type testCase struct {
		Description string
		Filter      asset.Filter
		Err         error
		Result      map[asset.Type]int
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	const (
		typeJob   = asset.Type("job")
		typeTable = asset.Type("table")
		typeTopic = asset.Type("topic")
	)

	testCases := []testCase{
		{
			Description: `should return error if asset repository get types return error`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetTypes(ctx, asset.Filter{}).Return(nil, errors.New("unknown error"))
			},
			Result: nil,
			Err:    errors.New("unknown error"),
		},
		{
			Description: `should return map types if asset repository get types return no error`,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetTypes(ctx, asset.Filter{}).Return(map[asset.Type]int{
					typeJob:   1,
					typeTable: 1,
					typeTopic: 1,
				}, nil)
			},
			Result: map[asset.Type]int{
				typeJob:   1,
				typeTable: 1,
				typeTopic: 1,
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{AssetRepo: mockAssetRepo})
			defer cancel()

			got, err := svc.GetTypes(ctx, tc.Filter)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
			if !cmp.Equal(tc.Result, got) {
				t.Fatalf("got result %+v, expected result was %+v", got, tc.Result)
			}
		})
	}
}

func TestService_UpsertAsset(t *testing.T) {
	sampleAsset := &asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.Type("dashboard"), Service: "some-service"}
	sampleNodes1 := []string{"1-urn-1", "1-urn-2"}
	sampleNodes2 := []string{"2-urn-1", "2-urn-2"}
	type testCase struct {
		Description string
		Asset       *asset.Asset
		Upstreams   []string
		Downstreams []string
		Err         error
		ReturnedID  string
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	testCases := []testCase{
		{
			Description: `should return error if asset repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(nil, errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: "",
		},
		{
			Description: `should return error if discovery repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: sampleAsset.ID,
		},
		{
			Description: `should return error if lineage repository upsert return error`,
			Asset:       sampleAsset,
			Upstreams:   sampleNodes1,
			Downstreams: sampleNodes2,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
				lr.EXPECT().Upsert(ctx, sampleAsset.URN, sampleNodes1, sampleNodes2).Return(errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: sampleAsset.ID,
		},
		{
			Description: `should return no error if all repositories upsert return no error`,
			Asset:       sampleAsset,
			Upstreams:   sampleNodes1,
			Downstreams: sampleNodes2,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
				lr.EXPECT().Upsert(ctx, sampleAsset.URN, sampleNodes1, sampleNodes2).Return(nil)
			},
			Err:        nil,
			ReturnedID: sampleAsset.ID,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			lineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, discoveryRepo, lineageRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   lineageRepo,
				Worker:        workermanager.NewInSituWorker(workermanager.Deps{DiscoveryRepo: discoveryRepo}),
			})
			defer cancel()

			rid, err := svc.UpsertAsset(ctx, tc.Asset, tc.Upstreams, tc.Downstreams)
			if tc.Err != nil {
				assert.ErrorContains(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.ReturnedID, rid)
		})
	}
}

func TestService_UpsertAssetWithoutLineage(t *testing.T) {
	sampleAsset := &asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.Type("dashboard"), Service: "some-service"}
	testCases := []struct {
		Description string
		Asset       *asset.Asset
		Err         error
		ReturnedID  string
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository)
	}{
		{
			Description: `should return error if asset repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(nil, errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return error if discovery repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return no error if all repositories upsert return no error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().Upsert(ctx, sampleAsset).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
			},
			ReturnedID: sampleAsset.ID,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, discoveryRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   mocks.NewLineageRepository(t),
				Worker:        workermanager.NewInSituWorker(workermanager.Deps{DiscoveryRepo: discoveryRepo}),
			})
			defer cancel()

			rid, err := svc.UpsertAssetWithoutLineage(ctx, tc.Asset)
			if tc.Err != nil {
				assert.ErrorContains(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.ReturnedID, rid)
		})
	}
}

func TestService_UpsertPatchAsset(t *testing.T) {
	sampleAsset := &asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.Type("dashboard"), Service: "some-service"}
	sampleNodes1 := []string{"1-urn-1", "1-urn-2"}
	sampleNodes2 := []string{"2-urn-1", "2-urn-2"}
	type testCase struct {
		Description string
		Asset       *asset.Asset
		Upstreams   []string
		Downstreams []string
		Err         error
		ReturnedID  string
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	testCases := []testCase{
		{
			Description: `should return error if asset repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.DiscoveryRepository, _ *mocks.LineageRepository) {
				ar.EXPECT().UpsertPatch(ctx, sampleAsset, mock.Anything).Return(nil, errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: "",
		},
		{
			Description: `should return error if discovery repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, _ *mocks.LineageRepository) {
				ar.EXPECT().UpsertPatch(ctx, sampleAsset, mock.Anything).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: sampleAsset.ID,
		},
		{
			Description: `should return error if lineage repository upsert return error`,
			Asset:       sampleAsset,
			Upstreams:   sampleNodes1,
			Downstreams: sampleNodes2,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().UpsertPatch(ctx, sampleAsset, mock.Anything).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
				lr.EXPECT().Upsert(ctx, sampleAsset.URN, sampleNodes1, sampleNodes2).Return(errors.New("unknown error"))
			},
			Err:        errors.New("unknown error"),
			ReturnedID: sampleAsset.ID,
		},
		{
			Description: `should return no error if all repositories upsert return no error`,
			Asset:       sampleAsset,
			Upstreams:   sampleNodes1,
			Downstreams: sampleNodes2,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().UpsertPatch(ctx, sampleAsset, mock.Anything).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
				lr.EXPECT().Upsert(ctx, sampleAsset.URN, sampleNodes1, sampleNodes2).Return(nil)
			},
			Err:        nil,
			ReturnedID: sampleAsset.ID,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			lineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, discoveryRepo, lineageRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   lineageRepo,
				Worker:        workermanager.NewInSituWorker(workermanager.Deps{DiscoveryRepo: discoveryRepo}),
			})
			defer cancel()

			patchData := make(map[string]interface{})
			rid, err := svc.UpsertPatchAsset(ctx, tc.Asset, tc.Upstreams, tc.Downstreams, patchData)
			if tc.Err != nil {
				assert.ErrorContains(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.ReturnedID, rid)
		})
	}
}

func TestService_UpsertPatchAssetWithoutLineage(t *testing.T) {
	sampleAsset := &asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.Type("dashboard"), Service: "some-service"}
	testCases := []struct {
		Description string
		Asset       *asset.Asset
		Err         error
		ReturnedID  string
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository)
	}{
		{
			Description: `should return error if asset repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.DiscoveryRepository) {
				ar.EXPECT().UpsertPatch(ctx, sampleAsset, mock.Anything).Return(nil, errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return error if discovery repository upsert return error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().UpsertPatch(ctx, sampleAsset, mock.Anything).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return no error if all repositories upsert return no error`,
			Asset:       sampleAsset,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository) {
				ar.EXPECT().UpsertPatch(ctx, sampleAsset, mock.Anything).Return(sampleAsset, nil)
				dr.EXPECT().Upsert(ctx, mock.AnythingOfType("asset.Asset")).Return(nil)
			},
			ReturnedID: sampleAsset.ID,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, discoveryRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   mocks.NewLineageRepository(t),
				Worker:        workermanager.NewInSituWorker(workermanager.Deps{DiscoveryRepo: discoveryRepo}),
			})
			defer cancel()

			patchData := make(map[string]interface{})
			rid, err := svc.UpsertPatchAssetWithoutLineage(ctx, tc.Asset, patchData)
			if tc.Err != nil {
				assert.ErrorContains(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.ReturnedID, rid)
		})
	}
}

func TestService_DeleteAsset(t *testing.T) {
	assetID := "d9351e2e-a6b2-4c5d-af68-b95432e30203"
	urn := "my-test-urn"
	type testCase struct {
		Description string
		ID          string
		Err         error
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	testCases := []testCase{
		{
			Description: `with ID, should return error if asset repository get by id returns error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByID(ctx, assetID).Return(urn, errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with ID, should return error if discovery repository delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByID(ctx, assetID).Return(urn, nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with ID, should return error if lineage repository delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByID(ctx, assetID).Return(urn, nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				lr.EXPECT().DeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with URN, should return error if asset repository delete return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with URN, should return error if discovery repository delete return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with URN, should return error if lineage repository delete return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				lr.EXPECT().DeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should call DeleteByURN on repositories by fetching URN when given an ID`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByID(ctx, assetID).Return(urn, nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				lr.EXPECT().DeleteByURN(ctx, urn).Return(nil)
			},
			Err: nil,
		},
		{
			Description: `should call DeleteByURN on repositories when not given an ID`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				dr.EXPECT().DeleteByURN(ctx, urn).Return(nil)
				lr.EXPECT().DeleteByURN(ctx, urn).Return(nil)
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			lineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, discoveryRepo, lineageRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   lineageRepo,
				Worker:        workermanager.NewInSituWorker(workermanager.Deps{DiscoveryRepo: discoveryRepo}),
			})
			defer cancel()

			err := svc.DeleteAsset(ctx, tc.ID)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
		})
	}
}

func TestService_SoftDeleteAsset(t *testing.T) {
	assetID := uuid.New().String()
	userID := uuid.New().String()
	urn := "my-test-urn"
	newVersion := "0.2"
	type testCase struct {
		Description string
		ID          string
		Err         error
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
	}

	testCases := []testCase{
		{
			Description: `with ID, should return error if asset repository soft delete by id returns error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.DiscoveryRepository, _ *mocks.LineageRepository) {
				ar.EXPECT().SoftDeleteByID(ctx, mock.AnythingOfType("time.Time"), assetID, userID).Return(urn, newVersion, errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with ID, should return error if discovery repository soft delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, _ *mocks.LineageRepository) {
				ar.EXPECT().SoftDeleteByID(ctx, mock.AnythingOfType("time.Time"), assetID, userID).Return(urn, newVersion, nil)
				dr.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("asset.SoftDeleteAssetParams")).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with URN, should return error if asset repository soft delete return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.DiscoveryRepository, _ *mocks.LineageRepository) {
				ar.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("time.Time"), urn, userID).Return(newVersion, errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `with URN, should return error if discovery repository soft delete return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, _ *mocks.LineageRepository) {
				ar.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("time.Time"), urn, userID).Return(newVersion, nil)
				dr.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("asset.SoftDeleteAssetParams")).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should return error if lineage repository soft delete return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().SoftDeleteByID(ctx, mock.AnythingOfType("time.Time"), assetID, userID).Return(urn, newVersion, nil)
				dr.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("asset.SoftDeleteAssetParams")).Return(nil)
				lr.EXPECT().SoftDeleteByURN(ctx, urn).Return(errors.New("unknown error"))
			},
			Err: errors.New("unknown error"),
		},
		{
			Description: `should call SoftDeleteByURN on repositories by fetching URN when given an ID`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().SoftDeleteByID(ctx, mock.AnythingOfType("time.Time"), assetID, userID).Return(urn, newVersion, nil)
				dr.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("asset.SoftDeleteAssetParams")).Return(nil)
				lr.EXPECT().SoftDeleteByURN(ctx, urn).Return(nil)
			},
			Err: nil,
		},
		{
			Description: `should call SoftDeleteByURN on repositories when not given an ID`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				ar.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("time.Time"), urn, userID).Return(newVersion, nil)
				dr.EXPECT().SoftDeleteByURN(ctx, mock.AnythingOfType("asset.SoftDeleteAssetParams")).Return(nil)
				lr.EXPECT().SoftDeleteByURN(ctx, urn).Return(nil)
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			lineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, discoveryRepo, lineageRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   lineageRepo,
				Worker:        workermanager.NewInSituWorker(workermanager.Deps{DiscoveryRepo: discoveryRepo}),
			})
			defer cancel()

			err := svc.SoftDeleteAsset(ctx, tc.ID, userID)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
		})
	}
}

func TestService_DeleteAssets(t *testing.T) {
	dummyRequestDryRunTrue := asset.DeleteAssetsRequest{
		QueryExpr: `testing < now()`,
		DryRun:    true,
	}
	dummyRequestDryRunFalse := asset.DeleteAssetsRequest{
		QueryExpr: `testing < now()`,
		DryRun:    false,
	}
	type testCase struct {
		Description        string
		Request            asset.DeleteAssetsRequest
		Setup              func(context.Context, *mocks.AssetRepository, *mocks.Worker, *mocks.LineageRepository)
		ExpectAffectedRows uint32
		ExpectErr          error
	}

	testCases := []testCase{
		{
			Description: `should return error if getting affected rows got error`,
			Request:     dummyRequestDryRunTrue,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.Worker, _ *mocks.LineageRepository) {
				ar.EXPECT().GetCountByQueryExpr(ctx, mock.AnythingOfType("asset.DeleteAssetExpr")).
					Return(0, errors.New("something wrong"))
			},
			ExpectAffectedRows: 0,
			ExpectErr:          errors.New("something wrong"),
		},
		{
			Description: `should only return the affected rows that match the given query when getting affected rows successful and dry run is true`,
			Request:     dummyRequestDryRunTrue,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.Worker, _ *mocks.LineageRepository) {
				ar.EXPECT().GetCountByQueryExpr(ctx, mock.AnythingOfType("asset.DeleteAssetExpr")).
					Return(11, nil)
			},
			ExpectAffectedRows: 11,
			ExpectErr:          nil,
		},
		{
			Description: `should return the affected rows and perform deletion in the background when getting affected rows successful and dry run is false`,
			Request:     dummyRequestDryRunFalse,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, w *mocks.Worker, lr *mocks.LineageRepository) {
				deletedURNs := []string{"urn1", "urn2"}
				ar.EXPECT().GetCountByQueryExpr(ctx, mock.AnythingOfType("asset.DeleteAssetExpr")).
					Return(2, nil)
				ar.EXPECT().DeleteByQueryExpr(mock.Anything, mock.Anything).
					Return(deletedURNs, nil)
				lr.EXPECT().DeleteByURNs(mock.Anything, mock.Anything).
					Return(nil)
				w.EXPECT().EnqueueDeleteAssetsByQueryExprJob(mock.Anything, mock.Anything).
					Return(nil)
			},
			ExpectAffectedRows: 2,
			ExpectErr:          nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			worker := mocks.NewWorker(t)
			lineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, worker, lineageRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   lineageRepo,
				Worker:        worker,
			})
			defer cancel()

			affectedRows, err := svc.DeleteAssets(ctx, tc.Request)
			time.Sleep(1 * time.Second)

			if tc.ExpectErr != nil {
				assert.ErrorContains(t, err, tc.ExpectErr.Error())
			}
			assert.Equal(t, tc.ExpectAffectedRows, affectedRows)
		})
	}
}

func TestService_SoftDeleteAssets(t *testing.T) {
	dummyRequestDryRunTrue := asset.DeleteAssetsRequest{
		QueryExpr: `testing < now()`,
		DryRun:    true,
	}
	dummyRequestDryRunFalse := asset.DeleteAssetsRequest{
		QueryExpr: `testing < now()`,
		DryRun:    false,
	}
	userID := uuid.New().String()
	type testCase struct {
		Description        string
		Request            asset.DeleteAssetsRequest
		Setup              func(context.Context, *mocks.AssetRepository, *mocks.Worker)
		ExpectAffectedRows uint32
		ExpectErr          error
	}

	testCases := []testCase{
		{
			Description: `should return error if getting affected rows got error`,
			Request:     dummyRequestDryRunTrue,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.Worker) {
				ar.EXPECT().GetCountByQueryExpr(ctx, mock.AnythingOfType("asset.DeleteAssetExpr")).
					Return(0, errors.New("something wrong"))
			},
			ExpectAffectedRows: 0,
			ExpectErr:          errors.New("something wrong"),
		},
		{
			Description: `should only return the affected rows that match the given query when getting affected rows successful and dry run is true`,
			Request:     dummyRequestDryRunTrue,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, _ *mocks.Worker) {
				ar.EXPECT().GetCountByQueryExpr(ctx, mock.AnythingOfType("asset.DeleteAssetExpr")).
					Return(11, nil)
			},
			ExpectAffectedRows: 11,
			ExpectErr:          nil,
		},
		{
			Description: `should return the affected rows and perform deletion in the background when getting affected rows successful and dry run is false`,
			Request:     dummyRequestDryRunFalse,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, w *mocks.Worker) {
				ar.EXPECT().GetCountByQueryExpr(ctx, mock.AnythingOfType("asset.DeleteAssetExpr")).
					Return(2, nil)
				ar.EXPECT().SoftDeleteByQueryExpr(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]asset.Asset{}, nil)
				w.EXPECT().EnqueueSoftDeleteAssetsJob(mock.Anything, mock.Anything).
					Return(nil)
			},
			ExpectAffectedRows: 2,
			ExpectErr:          nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			assetRepo := mocks.NewAssetRepository(t)
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			worker := mocks.NewWorker(t)
			lineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, assetRepo, worker)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     assetRepo,
				DiscoveryRepo: discoveryRepo,
				LineageRepo:   lineageRepo,
				Worker:        worker,
			})
			defer cancel()

			affectedRows, err := svc.SoftDeleteAssets(ctx, tc.Request, userID)
			time.Sleep(1 * time.Second)

			if tc.ExpectErr != nil {
				assert.ErrorContains(t, err, tc.ExpectErr.Error())
			}
			assert.Equal(t, tc.ExpectAffectedRows, affectedRows)
		})
	}
}

func TestService_GetAssetByID(t *testing.T) {
	assetID := "f742aa61-1100-445c-8d72-355a42e2fb59"
	urn := "my-test-urn"
	now := time.Now().UTC()
	type testCase struct {
		Description string
		ID          string
		Expected    *asset.Asset
		ExpectedErr error
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	ast := asset.Asset{
		ID: assetID,
	}

	testCases := []testCase{
		{
			Description: `should return error if the repository return error without id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{})
			},
			ExpectedErr: asset.NotFoundError{},
		},
		{
			Description: `should return error if the repository return error, with id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: ast.ID})
			},
			ExpectedErr: asset.NotFoundError{AssetID: ast.ID},
		},
		{
			Description: `should return error if the repository return error, with invalid id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.InvalidError{AssetID: ast.ID})
			},
			ExpectedErr: asset.InvalidError{AssetID: ast.ID},
		},
		{
			Description: `with URN, should return error from repository`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByURN(ctx, urn).Return(asset.Asset{}, errors.New("the world exploded"))
			},
			ExpectedErr: errors.New("the world exploded"),
		},
		{
			Description: `with ID, should return no error if asset is found`,
			ID:          assetID,
			Expected: &asset.Asset{
				ID:        assetID,
				URN:       urn,
				CreatedAt: now,
				Probes: []asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				},
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{
					ID:        assetID,
					URN:       urn,
					CreatedAt: now,
				}, nil)
				ar.EXPECT().GetProbes(ctx, urn).Return([]asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				}, nil)
			},
			ExpectedErr: nil,
		},
		{
			Description: `with URN, should return no error if asset is found`,
			ID:          urn,
			Expected: &asset.Asset{
				ID:        assetID,
				URN:       urn,
				CreatedAt: now,
				Probes: []asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				},
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByURN(ctx, urn).Return(asset.Asset{
					ID:        assetID,
					URN:       urn,
					CreatedAt: now,
				}, nil)
				ar.EXPECT().GetProbes(ctx, urn).Return([]asset.Probe{
					{ID: "probe-1", AssetURN: urn, Status: "RUNNING", Timestamp: now},
					{ID: "probe-2", AssetURN: urn, Status: "FAILED", Timestamp: now.Add(2 * time.Hour)},
				}, nil)
			},
			ExpectedErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     mockAssetRepo,
				DiscoveryRepo: mocks.NewDiscoveryRepository(t),
				LineageRepo:   mocks.NewLineageRepository(t),
			})
			defer cancel()

			actual, err := svc.GetAssetByID(ctx, tc.ID)
			if tc.Expected != nil {
				assert.Equal(t, *tc.Expected, actual)
			}
			if tc.ExpectedErr != nil {
				assert.ErrorContains(t, err, tc.ExpectedErr.Error())
				assert.ErrorAs(t, err, &tc.ExpectedErr)
			}
		})
	}
}

func TestService_GetAssetByIDWithoutProbes(t *testing.T) {
	assetID := "f742aa61-1100-445c-8d72-355a42e2fb59"
	urn := "my-test-urn"
	now := time.Now().UTC()
	ast := asset.Asset{
		ID: assetID,
	}

	testCases := []struct {
		Description string
		ID          string
		Expected    *asset.Asset
		ExpectedErr error
		Setup       func(context.Context, *mocks.AssetRepository)
	}{
		{
			Description: `should return error if the repository return error without id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{})
			},
			ExpectedErr: asset.NotFoundError{},
		},
		{
			Description: `should return error if the repository return error, with id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.NotFoundError{AssetID: ast.ID})
			},
			ExpectedErr: asset.NotFoundError{AssetID: ast.ID},
		},
		{
			Description: `should return error if the repository return error, with invalid id`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{}, asset.InvalidError{AssetID: ast.ID})
			},
			ExpectedErr: asset.InvalidError{AssetID: ast.ID},
		},
		{
			Description: `with URN, should return error from repository`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByURN(ctx, urn).Return(asset.Asset{}, errors.New("the world exploded"))
			},
			ExpectedErr: errors.New("the world exploded"),
		},
		{
			Description: `with ID, should return no error if asset is found`,
			ID:          assetID,
			Expected: &asset.Asset{
				ID:        assetID,
				URN:       urn,
				CreatedAt: now,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByID(ctx, assetID).Return(asset.Asset{
					ID:        assetID,
					URN:       urn,
					CreatedAt: now,
				}, nil)
			},
			ExpectedErr: nil,
		},
		{
			Description: `with URN, should return no error if asset is found`,
			ID:          urn,
			Expected: &asset.Asset{
				ID:        assetID,
				URN:       urn,
				CreatedAt: now,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByURN(ctx, urn).Return(asset.Asset{
					ID:        assetID,
					URN:       urn,
					CreatedAt: now,
				}, nil)
			},
			ExpectedErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     mockAssetRepo,
				DiscoveryRepo: mocks.NewDiscoveryRepository(t),
				LineageRepo:   mocks.NewLineageRepository(t),
			})
			defer cancel()

			actual, err := svc.GetAssetByIDWithoutProbes(ctx, tc.ID)
			if tc.Expected != nil {
				assert.Equal(t, *tc.Expected, actual)
			}
			if tc.ExpectedErr != nil {
				assert.ErrorContains(t, err, tc.ExpectedErr.Error())
				assert.ErrorAs(t, err, &tc.ExpectedErr)
			}
		})
	}
}

func TestService_GetAssetByVersion(t *testing.T) {
	assetID := "f742aa61-1100-445c-8d72-355a42e2fb59"
	urn := "my-test-urn"
	type testCase struct {
		Description string
		ID          string
		ExpectedErr error
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	testCases := []testCase{
		{
			Description: `should return error if the GetByVersionWithID function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithID(ctx, assetID, "v0.0.2").
					Return(asset.Asset{}, errors.New("error fetching asset"))
			},
			ExpectedErr: errors.New("error fetching asset"),
		},
		{
			Description: `should return error if the GetByVersionWithURN function return error`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithURN(ctx, urn, "v0.0.2").
					Return(asset.Asset{}, errors.New("error fetching asset"))
			},
			ExpectedErr: errors.New("error fetching asset"),
		},
		{
			Description: `should return no error if asset is found with ID`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithID(ctx, assetID, "v0.0.2").Return(asset.Asset{}, nil)
			},
			ExpectedErr: nil,
		},
		{
			Description: `should return no error if asset is found with URN`,
			ID:          urn,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetByVersionWithURN(ctx, urn, "v0.0.2").Return(asset.Asset{}, nil)
			},
			ExpectedErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     mockAssetRepo,
				DiscoveryRepo: mocks.NewDiscoveryRepository(t),
				LineageRepo:   mocks.NewLineageRepository(t),
			})
			defer cancel()

			_, err := svc.GetAssetByVersion(ctx, tc.ID, "v0.0.2")
			if tc.ExpectedErr != nil {
				assert.EqualError(t, err, tc.ExpectedErr.Error())
			}
		})
	}
}

func TestService_GetAssetVersionHistory(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		Err         error
		Setup       func(context.Context, *mocks.AssetRepository)
	}

	ast := []asset.Asset{}
	testCases := []testCase{
		{
			Description: `should return error if the GetVersionHistory function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{}, assetID).Return(ast, errors.New("error fetching asset"))
			},
			Err: errors.New("error fetching asset"),
		},
		{
			Description: `should return no error if asset is found by the version`,
			ID:          assetID,
			Setup: func(ctx context.Context, ar *mocks.AssetRepository) {
				ar.EXPECT().GetVersionHistory(ctx, asset.Filter{}, assetID).Return(ast, nil)
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     mockAssetRepo,
				DiscoveryRepo: mockDiscoveryRepo,
				LineageRepo:   mockLineageRepo,
			})
			defer cancel()

			_, err := svc.GetAssetVersionHistory(ctx, asset.Filter{}, tc.ID)
			if err != nil && errors.Is(tc.Err, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.Err)
			}
		})
	}
}

func TestService_GetLineage(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		Query       asset.LineageQuery
		Setup       func(context.Context, *mocks.AssetRepository, *mocks.DiscoveryRepository, *mocks.LineageRepository)
		Expected    asset.Lineage
		Err         error
	}

	testCases := []testCase{
		{
			Description: `should return error if the GetGraph function return error`,
			ID:          assetID,
			Query: asset.LineageQuery{
				WithAttributes: true,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{WithAttributes: true}).
					Return(asset.LineageGraph{}, errors.New("error fetching graph"))
			},
			Expected: asset.Lineage{},
			Err:      errors.New("error fetching graph"),
		},
		{
			Description: `should return no error if graph with 0 edges are returned`,
			ID:          assetID,
			Query: asset.LineageQuery{
				WithAttributes: true,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{WithAttributes: true}).
					Return(asset.LineageGraph{}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1"},
					MaxRows:   1,
				}).Return(nil, nil)
			},
			Expected: asset.Lineage{Edges: []asset.LineageEdge{}, NodeAttrs: map[string]asset.NodeAttributes{}},
			Err:      nil,
		},
		{
			Description: `should return an error if GetProbesWithFilter function returns error`,
			ID:          assetID,
			Query: asset.LineageQuery{
				WithAttributes: true,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{WithAttributes: true}).Return(asset.LineageGraph{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1", "urn-target-1", "urn-target-2", "urn-target-3"},
					MaxRows:   1,
				}).Return(nil, errors.New("error fetching probes"))
			},
			Expected: asset.Lineage{},
			Err:      errors.New("error fetching probes"),
		},
		{
			Description: `should return no error if GetProbesWithFilter function returns 0 probes`,
			ID:          assetID,
			Query: asset.LineageQuery{
				WithAttributes: true,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{WithAttributes: true}).Return(asset.LineageGraph{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1", "urn-target-1", "urn-target-2", "urn-target-3"},
					MaxRows:   1,
				}).Return(nil, nil)
			},
			Expected: asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				},
				NodeAttrs: map[string]asset.NodeAttributes{},
			},
			Err: nil,
		},
		{
			Description: `should return lineage with edges and without node attributes`,
			ID:          assetID,
			Query: asset.LineageQuery{
				WithAttributes: false,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{WithAttributes: false}).Return(asset.LineageGraph{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				}, nil)
			},
			Expected: asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				},
			},
			Err: nil,
		},
		{
			Description: `should return lineage with edges and node attributes`,
			ID:          assetID,
			Query: asset.LineageQuery{
				WithAttributes: true,
			},
			Setup: func(ctx context.Context, ar *mocks.AssetRepository, dr *mocks.DiscoveryRepository, lr *mocks.LineageRepository) {
				lr.EXPECT().GetGraph(ctx, "urn-source-1", asset.LineageQuery{WithAttributes: true}).Return(asset.LineageGraph{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				}, nil)
				ar.EXPECT().GetProbesWithFilter(ctx, asset.ProbesFilter{
					AssetURNs: []string{"urn-source-1", "urn-target-1", "urn-target-2", "urn-target-3"},
					MaxRows:   1,
				}).Return(
					map[string][]asset.Probe{
						"urn-source-1": {
							asset.Probe{Status: "SUCCESS"},
						},
						"urn-target-2": {},
						"urn-target-3": {
							asset.Probe{Status: "FAILED"},
						},
					},
					nil,
				)
			},
			Expected: asset.Lineage{
				Edges: []asset.LineageEdge{
					{Source: "urn-source-1", Target: "urn-target-1", Prop: nil},
					{Source: "urn-source-1", Target: "urn-target-2", Prop: nil},
					{Source: "urn-target-2", Target: "urn-target-3", Prop: nil},
				},
				NodeAttrs: map[string]asset.NodeAttributes{
					"urn-source-1": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "SUCCESS"},
						},
					},
					"urn-target-3": {
						Probes: asset.ProbesInfo{
							Latest: asset.Probe{Status: "FAILED"},
						},
					},
				},
			},
			Err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockAssetRepo, mockDiscoveryRepo, mockLineageRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     mockAssetRepo,
				DiscoveryRepo: mockDiscoveryRepo,
				LineageRepo:   mockLineageRepo,
			})
			defer cancel()

			actual, err := svc.GetLineage(ctx, "urn-source-1", tc.Query)
			if tc.Err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.Err.Error())
			}
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestService_SearchSuggestGroupAssets(t *testing.T) {
	assetID := "some-id"
	type testCase struct {
		Description string
		ID          string
		ErrSearch   error
		ErrSuggest  error
		ErrGroup    error
		Setup       func(context.Context, *mocks.DiscoveryRepository)
	}

	disErr := asset.DiscoveryError{
		Op:    "SearchSuggestGroupAssets",
		Err:   errors.New("could not find"),
		ID:    assetID,
		Index: "index",
	}

	searchResults := []asset.SearchResult{}
	groupResults := []asset.GroupResult{}
	testCases := []testCase{
		{
			Description: `should return error if the GetGraph function return error`,
			ID:          assetID,
			Setup: func(ctx context.Context, dr *mocks.DiscoveryRepository) {
				dr.EXPECT().Search(ctx, asset.SearchConfig{}).Return(searchResults, disErr)
				dr.EXPECT().Suggest(ctx, asset.SearchConfig{}).Return([]string{}, disErr)
				dr.EXPECT().GroupAssets(ctx, asset.GroupConfig{}).Return(groupResults, disErr)
			},
			ErrSearch:  disErr,
			ErrSuggest: disErr,
			ErrGroup:   disErr,
		},
		{
			Description: `should return no error if search, group  and suggest function work`,
			ID:          assetID,
			Setup: func(ctx context.Context, dr *mocks.DiscoveryRepository) {
				dr.EXPECT().Search(ctx, asset.SearchConfig{}).Return(searchResults, nil)
				dr.EXPECT().Suggest(ctx, asset.SearchConfig{}).Return([]string{}, nil)
				dr.EXPECT().GroupAssets(ctx, asset.GroupConfig{}).Return(groupResults, nil)
			},
			ErrSearch:  nil,
			ErrSuggest: nil,
			ErrGroup:   nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			mockAssetRepo := mocks.NewAssetRepository(t)
			mockDiscoveryRepo := mocks.NewDiscoveryRepository(t)
			mockLineageRepo := mocks.NewLineageRepository(t)
			if tc.Setup != nil {
				tc.Setup(ctx, mockDiscoveryRepo)
			}

			svc, cancel := asset.NewService(asset.ServiceDeps{
				AssetRepo:     mockAssetRepo,
				DiscoveryRepo: mockDiscoveryRepo,
				LineageRepo:   mockLineageRepo,
			})
			defer cancel()

			_, err := svc.SearchAssets(ctx, asset.SearchConfig{})
			if err != nil && !assert.Equal(t, tc.ErrSearch, err) {
				t.Fatalf("got error %v, expected error was %v", err, tc.ErrSearch)
			}
			_, err = svc.SuggestAssets(ctx, asset.SearchConfig{})
			if err != nil && !assert.Equal(t, tc.ErrSuggest.Error(), err.Error()) {
				t.Fatalf("got error %v, expected error was %v", err, tc.ErrSuggest)
			}

			_, err = svc.GroupAssets(ctx, asset.GroupConfig{})
			if err != nil && !assert.Equal(t, tc.ErrSuggest.Error(), err.Error()) {
				t.Fatalf("got error %v, expected error was %v", err, tc.ErrGroup)
			}
		})
	}
}

func TestService_CreateAssetProbe(t *testing.T) {
	var (
		ctx      = context.Background()
		assetURN = "sample-urn"
		probe    = asset.Probe{
			Status: "RUNNING",
		}
	)

	t.Run("should return no error on success", func(t *testing.T) {
		mockAssetRepo := mocks.NewAssetRepository(t)
		mockAssetRepo.EXPECT().AddProbe(ctx, assetURN, &probe).Return(nil)

		svc, cancel := asset.NewService(asset.ServiceDeps{AssetRepo: mockAssetRepo})
		defer cancel()

		err := svc.AddProbe(ctx, assetURN, &probe)
		assert.NoError(t, err)
	})

	t.Run("should return error on failed", func(t *testing.T) {
		expectedErr := errors.New("test error")

		mockAssetRepo := mocks.NewAssetRepository(t)
		mockAssetRepo.EXPECT().AddProbe(ctx, assetURN, &probe).Return(expectedErr)

		svc, cancel := asset.NewService(asset.ServiceDeps{AssetRepo: mockAssetRepo})
		defer cancel()

		err := svc.AddProbe(ctx, assetURN, &probe)
		assert.Equal(t, expectedErr, err)
	})
}
