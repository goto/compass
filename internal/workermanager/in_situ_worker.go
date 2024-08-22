package workermanager

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/pkg/queryexpr"
	"github.com/goto/salt/log"
)

type InSituWorker struct {
	discoveryRepo DiscoveryRepository
	assetRepo     asset.Repository
	mutex         sync.Mutex
	logger        log.Logger
}

func NewInSituWorker(deps Deps) *InSituWorker {
	return &InSituWorker{
		discoveryRepo: deps.DiscoveryRepo,
		assetRepo:     deps.AssetRepo,
		logger:        deps.Logger,
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

func (m *InSituWorker) EnqueueDeleteAssetsByQueryExprJob(ctx context.Context, queryExpr string) error {
	deleteESExpr := asset.DeleteAssetExpr{
		ExprStr: queryexpr.ESExpr(queryExpr),
	}
	if err := m.discoveryRepo.DeleteByQueryExpr(ctx, deleteESExpr); err != nil {
		return fmt.Errorf("delete asset from discovery repo: %w: query expr: '%s'", err, queryExpr)
	}
	return nil
}

func (m *InSituWorker) EnqueueSyncAssetJob(ctx context.Context, service string) error {
	const batchSize = 1000

	m.mutex.Lock()
	defer m.mutex.Unlock()

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

func (*InSituWorker) Close() error { return nil }
