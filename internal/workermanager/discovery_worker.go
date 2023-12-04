package workermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/pkg/worker"
)

//go:generate mockery --name=DiscoveryRepository -r --case underscore --with-expecter --structname DiscoveryRepository --filename discovery_repository_mock.go --output=./mocks

type DiscoveryRepository interface {
	Upsert(context.Context, asset.Asset) error
	DeleteByURN(ctx context.Context, assetURN string) error
	Clone(ctx context.Context, indexName, clonedIndexName string) error
	UpdateAlias(ctx context.Context, indexName, alias string) error
	DeleteByIndexName(ctx context.Context, indexName string) error
	UpdateIndexSettings(ctx context.Context, indexName string, body string) error
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
			Timeout:         5 * time.Second,
			BackoffStrategy: worker.DefaultExponentialBackoff,
		},
	}
}

func (m *Manager) syncAssetHandler() worker.JobHandler {
	return worker.JobHandler{
		Handle: m.SyncAssets,
		JobOpts: worker.JobOptions{
			MaxAttempts:     3,
			Timeout:         5 * time.Second,
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
	var service string
	if err := json.Unmarshal(job.Payload, &service); err != nil {
		return fmt.Errorf("sync asset: deserialise payload: %w", err)
	}

	jobs, err := m.jobRepo.GetSyncJobsByService(ctx, service)
	if err != nil {
		return fmt.Errorf("sync asset: get sync jobs by service: %w", err)
	}

	if len(jobs) > 1 {
		for _, job := range jobs {
			if job.RunAt.Before(job.RunAt) {
				return nil // mark job as done if there's earlier job with same service
			}
		}
	}

	backupIndexName := fmt.Sprintf("%+v-bak", service)

	err = m.discoveryRepo.Clone(ctx, service, backupIndexName)
	if err != nil {
		return fmt.Errorf("sync asset: clone index: %w", err)
	}

	err = m.discoveryRepo.UpdateAlias(ctx, backupIndexName, "universe")
	if err != nil {
		return fmt.Errorf("sync asset: update alias: %w", err)
	}

	err = m.discoveryRepo.DeleteByIndexName(ctx, service)
	if err != nil {
		return fmt.Errorf("sync asset: delete index: %w", err)
	}

	assets, err := m.assetRepo.GetAll(ctx, asset.Filter{
		Services: []string{service},
	})
	if err != nil {
		return fmt.Errorf("sync asset: get assets: %w", err)
	}

	for _, asset := range assets {
		err = m.discoveryRepo.Upsert(ctx, asset)
		if err != nil {
			return fmt.Errorf("sync asset: upsert assets in ES: %w", err)
		}
	}

	err = m.discoveryRepo.DeleteByIndexName(ctx, backupIndexName)
	if err != nil {
		return fmt.Errorf("sync asset: delete index: %w", err)
	}

	err = m.discoveryRepo.UpdateIndexSettings(ctx, service, `{"settings":{"index.blocks.write":false}}`)
	if err != nil {
		return fmt.Errorf("sync asset: update index settings: %w", err)
	}

	return nil
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
			Timeout:         5 * time.Second,
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
