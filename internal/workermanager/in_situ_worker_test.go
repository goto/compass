package workermanager_test

import (
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/core/asset"
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
	softDeleteAsset := asset.NewSoftDeleteAsset(currentTime, currentTime, "some-user")

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
			softDeleteAsset.URN = "some-urn"
			discoveryRepo.EXPECT().
				SoftDeleteByURN(ctx, softDeleteAsset).
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueSoftDeleteAssetJob(ctx, softDeleteAsset)
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

func TestInSituWorker_EnqueueSoftDeleteAssetsByQueryExprJob(t *testing.T) {
	currentTime := time.Now().UTC()
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
			softDeleteAssetsByQueryExpr := asset.NewSoftDeleteAssetsByQueryExpr(
				currentTime, currentTime, "some-user", queryExpr, deleteESExpr)

			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				SoftDeleteByQueryExpr(ctx, softDeleteAssetsByQueryExpr).
				Return(tc.discoveryErr)

			wrkr := workermanager.NewInSituWorker(workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := wrkr.EnqueueSoftDeleteAssetsByQueryExprJob(ctx, softDeleteAssetsByQueryExpr)
			if tc.expectedErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
