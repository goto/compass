package workermanager_test

import (
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/internal/workermanager/mocks"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/stretchr/testify/assert"
)

func TestInSituWorker_EnqueueIndexAssetJob(t *testing.T) {
	sampleAsset := asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.Type("dashboard"), Service: "some-service"}

	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "Failure",
			discoveryErr: errors.New("fail"),
			expectedErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				Upsert(ctx, sampleAsset).
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueIndexAssetJob(ctx, sampleAsset)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInSituWorker_EnqueueDeleteAssetJob(t *testing.T) {
	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "Failure",
			discoveryErr: errors.New("fail"),
			expectedErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				DeleteByURN(ctx, "some-urn").
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueDeleteAssetJob(ctx, "some-urn")
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInSituWorker_EnqueueSoftDeleteAssetJob(t *testing.T) {
	currentTime := time.Now().UTC()
	params := asset.SoftDeleteAssetParams{
		URN:         "some-urn",
		UpdatedAt:   currentTime,
		RefreshedAt: currentTime,
		NewVersion:  "0.1",
		UpdatedBy:   "some-user",
	}

	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "Failure",
			discoveryErr: errors.New("fail"),
			expectedErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				SoftDeleteByURN(ctx, params).
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueSoftDeleteAssetJob(ctx, params)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInSituWorker_EnqueueDeleteAssetsByQueryExprJob(t *testing.T) {
	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "Failure",
			discoveryErr: errors.New("fail"),
			expectedErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queryExpr := "refreshed_at <= '" + time.Now().Format("2006-01-02T15:04:05Z") +
				"' && service == 'test-service'" +
				"' && type == 'table'" +
				"' && urn == 'some-urn'"
			deleteESExpr := asset.DeleteAssetExpr{
				ExprStr: queryexpr.ESExpr(queryExpr),
			}
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				DeleteByQueryExpr(ctx, deleteESExpr).
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueDeleteAssetsByQueryExprJob(ctx, queryExpr)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInSituWorker_EnqueueSoftDeleteAssetsJob(t *testing.T) {
	currentTime := time.Now().UTC()
	dummyAssets := []asset.Asset{
		{
			ID:          "asset-1",
			URN:         "urn:asset:1",
			Type:        asset.Type("table"),
			Service:     "test-service",
			UpdatedAt:   currentTime,
			RefreshedAt: &currentTime,
			UpdatedBy:   user.User{ID: "some-user"},
		},
	}
	cases := []struct {
		discoveryErr error
		name         string
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "Failure",
			discoveryErr: errors.New("fail"),
			expectedErr:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				SoftDeleteAssets(ctx, dummyAssets, false).
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueSoftDeleteAssetsJob(ctx, dummyAssets)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
