package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/internal/cleanup"
	"github.com/goto/compass/internal/store/elasticsearch"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/pkg/telemetry"
	"github.com/goto/salt/term"
	"github.com/spf13/cobra"
)

func cleanupCmd(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup <command>",
		Short: "Run compass cleanup",
		Long:  "Compass cleanup management commands",
		Example: heredoc.Doc(`
			$ compass cleanup start
			$ compass cleanup start -c ./config.yaml
		`),
	}

	cmd.AddCommand(cleanupStartCommand(cfg))

	return cmd
}

func cleanupStartCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run compass cleanup",
		Long:  "Compass start cleanup management commands",
		Example: heredoc.Doc(`
			$ compass cleanup start
			$ compass cleanup start -c ./config.yaml
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			total, err := runCleanUp(cmd.Context(), cfg)
			if err != nil {
				return fmt.Errorf("run cleanup: %w", err)
			}

			fmt.Println("Compass cleanup completed successfully", term.Yellowf("with total deleted assets %v", total))
			return nil
		},
	}

	return cmd
}

func runCleanUp(ctx context.Context, cfg *Config) (uint32, error) {
	logger := initLogger(cfg.LogLevel)
	logger.Info("Compass cleanup starting", "version", Version)

	_, otelCleanup, err := telemetry.Init(ctx, cfg.Telemetry, logger)
	if err != nil {
		return 0, err
	}

	defer otelCleanup()

	esClient, err := initElasticsearch(logger, cfg.Elasticsearch)
	if err != nil {
		return 0, err
	}

	pgClient, err := initPostgres(ctx, logger, cfg)
	if err != nil {
		return 0, err
	}

	// Initialize repositories
	userRepository, err := postgres.NewUserRepository(pgClient)
	if err != nil {
		return 0, fmt.Errorf("create new user repository: %w", err)
	}
	assetRepository, err := postgres.NewAssetRepository(pgClient, userRepository, 0, cfg.Service.Identity.ProviderDefaultName)
	if err != nil {
		return 0, fmt.Errorf("create new asset repository: %w", err)
	}
	discoveryRepository := elasticsearch.NewDiscoveryRepository(
		esClient,
		logger,
		cfg.Elasticsearch.RequestTimeout,
		strings.Split(cfg.ColSearchExclusionKeywords, ","))
	lineageRepository, err := postgres.NewLineageRepository(pgClient)
	if err != nil {
		return 0, fmt.Errorf("create new lineage repository: %w", err)
	}

	wrkr, err := initAssetWorker(ctx, workermanager.Deps{
		Config:        cfg.Worker,
		DiscoveryRepo: discoveryRepository,
		AssetRepo:     assetRepository,
		Logger:        logger,
	})
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := wrkr.Close(); err != nil {
			logger.Error("Close worker", "err", err)
		}
	}()

	assetService, cancel := asset.NewService(asset.ServiceDeps{
		AssetRepo:     assetRepository,
		DiscoveryRepo: discoveryRepository,
		LineageRepo:   lineageRepository,
		Worker:        wrkr,
		Logger:        logger,
		Config:        cfg.Asset,
	})
	defer cancel()

	return cleanup.Run(ctx, cfg.Cleanup, assetService)
}
