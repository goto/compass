package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/goto/compass/core/asset"
	"github.com/goto/compass/core/discussion"
	"github.com/goto/compass/core/star"
	"github.com/goto/compass/core/tag"
	"github.com/goto/compass/core/user"
	compassserver "github.com/goto/compass/internal/server"
	esStore "github.com/goto/compass/internal/store/elasticsearch"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/workermanager"
	"github.com/goto/compass/pkg/telemetry"
	"github.com/goto/salt/log"
	"github.com/spf13/cobra"
)

// Version of the current build. overridden by the build system.
// see "Makefile" for more information
var (
	Version string
)

func serverCmd(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server <command>",
		Aliases: []string{"s"},
		Short:   "Run compass server",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ compass server start
			$ compass server start -c ./config.yaml
			$ compass server migrate
			$ compass server migrate -c ./config.yaml
		`),
	}

	cmd.AddCommand(
		serverStartCommand(cfg),
		serverMigrateCommand(cfg),
	)

	return cmd
}

func serverStartCommand(cfg *Config) *cobra.Command {
	c := &cobra.Command{
		Use:     "start",
		Short:   "Start server on default port 8080",
		Example: "compass server start",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runServer(cmd.Context(), cfg); err != nil {
				return fmt.Errorf("run server: %w", err)
			}
			return nil
		},
	}

	return c
}

func serverMigrateCommand(cfg *Config) *cobra.Command {
	c := &cobra.Command{
		Use:   "migrate",
		Short: "Run storage migration",
		Example: heredoc.Doc(`
			$ compass server migrate
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrations(cmd.Context(), cfg)
		},
	}

	return c
}

func runServer(ctx context.Context, cfg *Config) error {
	logger := initLogger(cfg.LogLevel)
	logger.Info("compass starting", "version", Version)

	nrApp, cleanUp, err := telemetry.Init(ctx, cfg.Telemetry, logger)
	if err != nil {
		return err
	}

	defer cleanUp()

	esClient, err := initElasticsearch(logger, cfg.Elasticsearch)
	if err != nil {
		return err
	}

	pgClient, err := initPostgres(ctx, logger, cfg)
	if err != nil {
		return err
	}

	// init tag
	tagRepository, err := postgres.NewTagRepository(pgClient)
	if err != nil {
		return fmt.Errorf("create new tag repository: %w", err)
	}
	tagTemplateRepository, err := postgres.NewTagTemplateRepository(pgClient)
	if err != nil {
		return fmt.Errorf("create new tag template repository: %w", err)
	}
	tagTemplateService := tag.NewTemplateService(tagTemplateRepository)
	tagService := tag.NewService(tagRepository, tagTemplateService)

	// init user
	userRepository, err := postgres.NewUserRepository(pgClient)
	if err != nil {
		return fmt.Errorf("create new user repository: %w", err)
	}
	userService := user.NewService(logger, userRepository)

	assetRepository, err := postgres.NewAssetRepository(pgClient, userRepository, 0, cfg.Service.Identity.ProviderDefaultName)
	if err != nil {
		return fmt.Errorf("create new asset repository: %w", err)
	}
	discoveryRepository := esStore.NewDiscoveryRepository(esClient, logger, cfg.Elasticsearch.RequestTimeout, strings.Split(cfg.ColSearchExclusionKeywords, ","))
	lineageRepository, err := postgres.NewLineageRepository(pgClient)
	if err != nil {
		return fmt.Errorf("create new lineage repository: %w", err)
	}

	wrkr, err := initAssetWorker(ctx, workermanager.Deps{
		Config:        cfg.Worker,
		DiscoveryRepo: discoveryRepository,
		AssetRepo:     assetRepository,
		Logger:        logger,
	})
	if err != nil {
		return err
	}

	defer func() {
		if err := wrkr.Close(); err != nil {
			logger.Error("Close worker", "err", err)
		}
	}()

	assetService := asset.NewService(asset.ServiceDeps{
		AssetRepo:     assetRepository,
		DiscoveryRepo: discoveryRepository,
		LineageRepo:   lineageRepository,
		Worker:        wrkr,
		Logger:        logger,
	})

	// init discussion
	discussionRepository, err := postgres.NewDiscussionRepository(pgClient, 0)
	if err != nil {
		return fmt.Errorf("create new discussion repository: %w", err)
	}
	discussionService := discussion.NewService(discussionRepository)

	// init star
	starRepository, err := postgres.NewStarRepository(pgClient)
	if err != nil {
		return fmt.Errorf("create new star repository: %w", err)
	}
	starService := star.NewService(starRepository)

	return compassserver.Serve(
		ctx,
		cfg.Service,
		logger,
		pgClient,
		nrApp,
		assetService,
		starService,
		discussionService,
		tagService,
		tagTemplateService,
		userService,
	)
}

func initLogger(logLevel string) *log.Logrus {
	logger := log.NewLogrus(
		log.LogrusWithLevel(logLevel),
		log.LogrusWithWriter(os.Stdout),
	)
	return logger
}

func initElasticsearch(logger log.Logger, config esStore.Config) (*esStore.Client, error) {
	esClient, err := esStore.NewClient(logger, config)
	if err != nil {
		return nil, fmt.Errorf("create new elasticsearch client: %w", err)
	}
	got, err := esClient.Init()
	if err != nil {
		return nil, fmt.Errorf("establish connection to elasticsearch: %w", err)
	}
	logger.Info("connected to elasticsearch", "info", got)
	return esClient, nil
}

func initPostgres(ctx context.Context, logger log.Logger, config *Config) (*postgres.Client, error) {
	pgClient, err := postgres.NewClient(ctx, config.DB)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres client: %w", err)
	}
	logger.Info("connected to postgres server", "host", config.DB.Host, "port", config.DB.Port)

	return pgClient, nil
}

func initAssetWorker(ctx context.Context, deps workermanager.Deps) (asset.Worker, error) {
	if !deps.Config.Enabled {
		return workermanager.NewInSituWorker(deps), nil
	}

	mgr, err := workermanager.New(ctx, deps)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func runMigrations(ctx context.Context, config *Config) error {
	fmt.Println("Preparing migration...")

	logger := initLogger(config.LogLevel)
	logger.Info("compass is migrating", "version", Version)

	logger.Info("Migrating Postgres...")
	if err := migratePostgres(ctx, logger, config); err != nil {
		return err
	}
	logger.Info("Migration Postgres done.")

	return nil
}

func migratePostgres(ctx context.Context, logger log.Logger, config *Config) (err error) {
	logger.Info("Initiating Postgres client...")

	pgClient, err := postgres.NewClient(ctx, config.DB)
	if err != nil {
		logger.Error("failed to prepare migration", "error", err)
		return err
	}

	err = pgClient.Migrate()
	if err != nil {
		return fmt.Errorf("problem with migration %w", err)
	}

	return nil
}
