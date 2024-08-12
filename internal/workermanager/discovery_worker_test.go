package workermanager_test

import (
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/testutils"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/internal/workermanager/mocks"
	"github.com/goto/compass/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func TestManager_EnqueueIndexAssetJob(t *testing.T) {
	sampleAsset := asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.TypeDashboard, Service: "some-service"}

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
					Type:    workermanager.JobIndexAsset,
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
	sampleAsset := asset.Asset{ID: "some-id", URN: "some-urn", Type: asset.TypeDashboard, Service: "some-service"}

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
				Type:    workermanager.JobIndexAsset,
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
					Type:    workermanager.JobDeleteAsset,
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
				Type:    workermanager.JobDeleteAsset,
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
			queryExpr := "refreshed_at <= '" + time.Now().Format("2006-01-02 15:04:05") +
				"' && service == 'test-service'" +
				"' && type == 'table'" +
				"' && urn == 'some-urn'"
			wrkr := mocks.NewWorker(t)
			wrkr.EXPECT().
				Enqueue(ctx, worker.JobSpec{
					Type:    workermanager.JobDeleteAssetsByQuery,
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
			queryExpr := "refreshed_at <= '" + time.Now().Format("2006-01-02 15:04:05") +
				"' && service == 'test-service'" +
				"' && type == 'table'" +
				"' && urn == 'some-urn'"
			discoveryRepo := mocks.NewDiscoveryRepository(t)
			discoveryRepo.EXPECT().
				DeleteByQueryExpr(ctx, queryExpr).
				Return(tc.discoveryErr)

			mgr := workermanager.NewWithWorker(mocks.NewWorker(t), workermanager.Deps{
				DiscoveryRepo: discoveryRepo,
			})
			err := mgr.DeleteAssetsByQueryExpr(ctx, worker.JobSpec{
				Type:    workermanager.JobDeleteAssetsByQuery,
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
