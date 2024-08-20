package workermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/pkg/worker"
)

//go:generate mockery --name=DiscoveryRepository -r --case underscore --with-expecter --structname DiscoveryRepository --filename discovery_repository_mock.go --output=./mocks

type DiscoveryRepository interface {
	Upsert(context.Context, asset.Asset) error
	DeleteByURN(ctx context.Context, assetURN string) error
	DeleteByQueryExpr(ctx context.Context, queryExpr string) error
	SyncAssets(ctx context.Context, indexName string) (cleanupFn func() error, err error)
}

func (m *Manager) EnqueueIndexAssetJob(ctx context.Context, ast asset.Asset) error {
	payload, err := json.Marshal(ast)
	if err != nil {
		return fmt.Errorf("enqueue index asset job: serialize payload: %w", err)
	}

	err = m.worker.Enqueue(ctx, worker.JobSpec{
		Type:    jobIndexAsset,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("enqueue index asset job: %w", err)
	}

	return nil
}

func (m *Manager) indexAssetHandler() worker.JobHandler {
	return worker.JobHandler{
		Handle: m.IndexAsset,
		JobOpts: worker.JobOptions{
			MaxAttempts:     3,
			Timeout:         m.indexTimeout,
			BackoffStrategy: worker.DefaultExponentialBackoff,
		},
	}
}

func (m *Manager) syncAssetHandler() worker.JobHandler {
	return worker.JobHandler{
		Handle: m.SyncAssets,
		JobOpts: worker.JobOptions{
			MaxAttempts:     1,
			Timeout:         m.syncTimeout,
			BackoffStrategy: worker.DefaultExponentialBackoff,
		},
	}
}

func (m *Manager) IndexAsset(ctx context.Context, job worker.JobSpec) error {
	var ast asset.Asset
	if err := json.Unmarshal(job.Payload, &ast); err != nil {
		return fmt.Errorf("index asset: deserialise payload: %w", err)
	}

	if err := m.discoveryRepo.Upsert(ctx, ast); err != nil {
		return &worker.RetryableError{
			Cause: fmt.Errorf("index asset: upsert into discovery repo: %w: urn '%s'", err, ast.URN),
		}
	}
	return nil
}

func (m *Manager) SyncAssets(ctx context.Context, job worker.JobSpec) error {
	const batchSize = 1000
	service := string(job.Payload)

	jobs, err := m.worker.GetSyncJobsByService(ctx, service)
	if err != nil {
		return fmt.Errorf("sync asset: get sync jobs by service: %w", err)
	}

	if len(jobs) > 1 {
		for _, job := range jobs {
			if job.RunAt.Before(job.RunAt) {
				return nil // mark job as done if there's earlier job with same service to prevent race conditions
			}
		}
	}

	cleanupFn, err := m.discoveryRepo.SyncAssets(ctx, service)
	if err != nil {
		return err
	}

	it := 0

	for {
		assets, err := m.assetRepo.GetAll(ctx, asset.Filter{
			Services: []string{service},
			Size:     batchSize,
			Offset:   it * batchSize,
			SortBy:   "name",
		})
		if err != nil {
			return fmt.Errorf("sync asset: get assets: %w", err)
		}

		for _, ast := range assets {
			if err := m.discoveryRepo.Upsert(ctx, ast); err != nil {
				if strings.Contains(err.Error(), "illegal_argument_exception") {
					m.logger.Error(err.Error())
					continue
				}
				return err
			}
		}

		if len(assets) != batchSize {
			break
		}
		it++
	}

	return cleanupFn()
}

func (m *Manager) EnqueueDeleteAssetJob(ctx context.Context, urn string) error {
	err := m.worker.Enqueue(ctx, worker.JobSpec{
		Type:    jobDeleteAsset,
		Payload: ([]byte)(urn),
	})
	if err != nil {
		return fmt.Errorf("enqueue delete asset job: %w: urn '%s'", err, urn)
	}

	return nil
}

func (m *Manager) deleteAssetHandler() worker.JobHandler {
	return worker.JobHandler{
		Handle: m.DeleteAsset,
		JobOpts: worker.JobOptions{
			MaxAttempts:     3,
			Timeout:         m.deleteTimeout,
			BackoffStrategy: worker.DefaultExponentialBackoff,
		},
	}
}

func (m *Manager) DeleteAsset(ctx context.Context, job worker.JobSpec) error {
	urn := (string)(job.Payload)
	if err := m.discoveryRepo.DeleteByURN(ctx, urn); err != nil {
		return &worker.RetryableError{
			Cause: fmt.Errorf("delete asset from discovery repo: %w: urn '%s'", err, urn),
		}
	}
	return nil
}

func (m *Manager) EnqueueDeleteAssetsByQueryExprJob(ctx context.Context, queryExpr string) error {
	err := m.worker.Enqueue(ctx, worker.JobSpec{
		Type:    jobDeleteAssetsByQuery,
		Payload: ([]byte)(queryExpr),
	})
	if err != nil {
		return fmt.Errorf("enqueue delete asset job: %w: query expr: '%s'", err, queryExpr)
	}

	return nil
}

func (m *Manager) deleteAssetsByQueryHandler() worker.JobHandler {
	return worker.JobHandler{
		Handle: m.DeleteAssetsByQueryExpr,
		JobOpts: worker.JobOptions{
			MaxAttempts:     3,
			Timeout:         m.indexTimeout,
			BackoffStrategy: worker.DefaultExponentialBackoff,
		},
	}
}

func (m *Manager) DeleteAssetsByQueryExpr(ctx context.Context, job worker.JobSpec) error {
	queryExpr := (string)(job.Payload)
	if err := m.discoveryRepo.DeleteByQueryExpr(ctx, queryExpr); err != nil {
		return &worker.RetryableError{
			Cause: fmt.Errorf("delete asset from discovery repo: %w: query expr: '%s'", err, queryExpr),
		}
	}
	return nil
}

func (m *Manager) EnqueueSyncAssetJob(ctx context.Context, service string) error {
	err := m.worker.Enqueue(ctx, worker.JobSpec{
		Type:    jobSyncAsset,
		Payload: ([]byte)(service),
	})
	if err != nil {
		return fmt.Errorf("enqueue sync asset job: %w: service '%s'", err, service)
	}

	return nil
}
