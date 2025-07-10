package cleanup

import (
	"context"
	"fmt"

	handlersv1beta1 "github.com/goto/compass/internal/server/v1beta1"
)

func Run(ctx context.Context, cfg Config, assetService handlersv1beta1.AssetService) (uint32, error) {
	deletedCount, err := assetService.DeleteAssetsByServicesAndUpdatedAt(ctx, cfg.DryRun, cfg.Services, cfg.ExpiryTime)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup assets: %w", err)
	}

	return deletedCount, nil
}
