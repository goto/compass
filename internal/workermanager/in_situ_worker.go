package workermanager

import (
	"context"
	"fmt"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/job"
)

type InSituWorker struct {
	discoveryRepo DiscoveryRepository
	jobRepo       job.Repository
	assetRepo     asset.Repository
}

func NewInSituWorker(deps Deps) *InSituWorker {
	return &InSituWorker{
		discoveryRepo: deps.DiscoveryRepo,
	}
}

func (m *InSituWorker) EnqueueIndexAssetJob(ctx context.Context, ast asset.Asset) error {
	if err := m.discoveryRepo.Upsert(ctx, ast); err != nil {
		return fmt.Errorf("index asset: upsert into discovery repo: %w: urn '%s'", err, ast.URN)
	}

	return nil
}

func (m *InSituWorker) EnqueueDeleteAssetJob(ctx context.Context, urn string) error {
	if err := m.discoveryRepo.DeleteByURN(ctx, urn); err != nil {
		return fmt.Errorf("delete asset from discovery repo: %w: urn '%s'", err, urn)
	}
	return nil
}

func (m *InSituWorker) EnqueueSyncAssetJob(ctx context.Context, service string) error {
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

func (*InSituWorker) Close() error { return nil }
