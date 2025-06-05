package workermanager_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/testutils"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/internal/workermanager/mocks"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/goto/compass/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func TestManager_EnqueueIndexAssetJob(t *testing.T) {
	sampleAsset := asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.Type("dashboard"), Service: "some-service"}

	cases := []struct {
		name        string
		enqueueErr  error
		expectedErr string
	}{
		{name: "Success"},
		{
			name:        "Failure",
			enqueueErr:  errors.New("fail"),
			expectedErr: "enqueue index asset job: fail",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wrkr := mocks.NewWorker(t)
			wrkr.EXPECT().
				Enqueue(ctx, worker.JobSpec{
					Type:    "index-asset",
					Payload: testutils.Marshal(t, sampleAsset),
				}).
				Return(tc.enqueueErr)

			mgr := workermanager.NewWithWorker(wrkr, workermanager.Deps{})
			err := mgr.EnqueueIndexAssetJob(ctx, sampleAsset)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_IndexAsset(t *testing.T) {
	sampleAsset := asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.Type("dashboard"), Service: "some-service"}

	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "failure",
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

			mgr := workermanager.NewWithWorker(mocks.NewWorker(t), workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := mgr.IndexAsset(ctx, worker.JobSpec{
				Type:    "index-asset",
				Payload: testutils.Marshal(t, sampleAsset),
			})
			if tc.expectedErr {
				var re *worker.RetryableError
				assert.ErrorAs(t, err, &re)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_EnqueueDeleteAssetJob(t *testing.T) {
	cases := []struct {
		name        string
		enqueueErr  error
		expectedErr string
	}{
		{name: "Success"},
		{
			name:        "Failure",
			enqueueErr:  errors.New("fail"),
			expectedErr: "enqueue delete asset job: fail",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wrkr := mocks.NewWorker(t)
			wrkr.EXPECT().
				Enqueue(ctx, worker.JobSpec{
					Type:    "delete-asset",
					Payload: []byte("some-urn"),
				}).
				Return(tc.enqueueErr)

			mgr := workermanager.NewWithWorker(wrkr, workermanager.Deps{})
			err := mgr.EnqueueDeleteAssetJob(ctx, "some-urn")
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_DeleteAsset(t *testing.T) {
	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "failure",
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

			mgr := workermanager.NewWithWorker(mocks.NewWorker(t), workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := mgr.DeleteAsset(ctx, worker.JobSpec{
				Type:    "delete-asset",
				Payload: []byte("some-urn"),
			})
			if tc.expectedErr {
				var re *worker.RetryableError
				assert.ErrorAs(t, err, &re)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_EnqueueSoftDeleteAssetJob(t *testing.T) {
	currentTime := time.Now().UTC()
	params := asset.SoftDeleteAssetParams{
		URN:         "some-urn",
		UpdatedAt:   currentTime,
		RefreshedAt: currentTime,
		NewVersion:  "0.1",
		UpdatedBy:   "some-user",
	}

	cases := []struct {
		name        string
		enqueueErr  error
		expectedErr string
	}{
		{name: "Success"},
		{
			name:        "Failure",
			enqueueErr:  errors.New("fail"),
			expectedErr: "enqueue soft delete asset job: fail",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wrkr := mocks.NewWorker(t)
			payload, err := json.Marshal(params)
			assert.NoError(t, err, "failed to marshal soft delete asset")
			wrkr.EXPECT().
				Enqueue(ctx, worker.JobSpec{
					Type:    "soft-delete-asset",
					Payload: payload,
				}).
				Return(tc.enqueueErr)

			mgr := workermanager.NewWithWorker(wrkr, workermanager.Deps{})
			err = mgr.EnqueueSoftDeleteAssetJob(ctx, params)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_SoftDeleteAsset(t *testing.T) {
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
			name:         "failure",
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

			mgr := workermanager.NewWithWorker(mocks.NewWorker(t), workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})

			payload, err := json.Marshal(params)
			assert.NoError(t, err, "failed to marshal soft delete asset")
			err = mgr.SoftDeleteAsset(ctx, worker.JobSpec{
				Type:    "soft-delete-asset",
				Payload: payload,
			})
			if tc.expectedErr {
				var re *worker.RetryableError
				assert.ErrorAs(t, err, &re)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_EnqueueDeleteAssetsByQueryExprJob(t *testing.T) {
	cases := []struct {
		name        string
		enqueueErr  error
		expectedErr string
	}{
		{name: "Success"},
		{
			name:        "Failure",
			enqueueErr:  errors.New("fail"),
			expectedErr: "enqueue delete asset job: fail: query expr:",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queryExpr := "refreshed_at <= '" + time.Now().Format("2006-01-02T15:04:05Z") +
				"' && service == 'test-service'" +
				"' && type == 'table'" +
				"' && urn == 'some-urn'"
			wrkr := mocks.NewWorker(t)
			wrkr.EXPECT().
				Enqueue(ctx, worker.JobSpec{
					Type:    "delete-assets-by-query",
					Payload: []byte(queryExpr),
				}).
				Return(tc.enqueueErr)

			mgr := workermanager.NewWithWorker(wrkr, workermanager.Deps{})
			err := mgr.EnqueueDeleteAssetsByQueryExprJob(ctx, queryExpr)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_DeleteAssets(t *testing.T) {
	cases := []struct {
		name         string
		discoveryErr error
		expectedErr  bool
	}{
		{name: "Success"},
		{
			name:         "failure",
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

			mgr := workermanager.NewWithWorker(mocks.NewWorker(t), workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := mgr.DeleteAssetsByQueryExpr(ctx, worker.JobSpec{
				Type:    "delete-assets-by-query",
				Payload: []byte(queryExpr),
			})
			if tc.expectedErr {
				var re *worker.RetryableError
				assert.ErrorAs(t, err, &re)
				assert.ErrorIs(t, err, tc.discoveryErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
