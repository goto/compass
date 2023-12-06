package workermanager

import (
	"context"
	"fmt"
	"time"

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
		jobRepo:       deps.JobRepo,
		assetRepo:     deps.AssetRepo,
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
	const batchSize = 1000
	jobs, err := m.jobRepo.GetSyncJobsByService(ctx, service)
	if err != nil {
		return fmt.Errorf("sync asset: get sync jobs by service: %w", err)
	}

	if len(jobs) > 0 {
		return nil // mark job as done if there's earlier job with same service
	}

	jobID, err := m.jobRepo.Insert(ctx, jobSyncAsset, ([]byte)(service), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("sync asset: insert job queue: %w", err)
	}
	defer func() {
		_ = m.jobRepo.Delete(ctx, jobID)
	}()

	cleanup, err := m.discoveryRepo.SyncAssets(ctx, service)
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
				return err
			}
		}

		if len(assets) != batchSize {
			break
		}
		it++
	}

	return cleanup()
}

func (*InSituWorker) Close() error { return nil }
