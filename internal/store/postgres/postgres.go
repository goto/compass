//go:build go1.16
// +build go1.16

package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	// Register database postgres
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Register golang migrate source
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpgx" // register instrumented DB driver
	"go.nhat.io/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

//go:embed migrations/*.sql
var fs embed.FS

const (
	columnNameCreatedAt     = "created_at"
	columnNameUpdatedAt     = "updated_at"
	sortDirectionAscending  = "ASC"
	sortDirectionDescending = "DESC"
	DEFAULT_MAX_RESULT_SIZE = 100
)

type Client struct {
	db *sqlx.DB
}

func (c *Client) RunWithinTx(ctx context.Context, f func(tx *sqlx.Tx) error) error {
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	if err := f(tx); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			//nolint:errorlint
			return fmt.Errorf("rollback transaction error: %v (original error: %w)", txErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (c *Client) Migrate() (err error) {
	m, err := initMigration(c.db.DB)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

// ExecQueries is used for executing list of db query
func (c *Client) ExecQueries(ctx context.Context, queries []string) error {
	for _, query := range queries {
		_, err := c.db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

// NewClient initializes database connection
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	driverName, err := otelsql.Register(
		"nrpgx",
		otelsql.TraceQueryWithoutArgs(),
		otelsql.TraceRowsClose(),
		otelsql.TraceRowsAffected(),
		otelsql.WithSystem(semconv.DBSystemPostgreSQL),
		otelsql.WithInstanceName("default"),
	)
	if err != nil {
		return nil, fmt.Errorf("register otelsql: %w", err)
	}

	db, err := sql.Open(driverName, cfg.ConnectionURL().String())
	if err != nil {
		return nil, fmt.Errorf("open DB connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping DB: %w", err)
	}

	if err := otelsql.RecordStats(
		db,
		otelsql.WithSystem(semconv.DBSystemPostgreSQL),
		otelsql.WithInstanceName("default"),
	); err != nil {
		return nil, err
	}

	return &Client{db: sqlx.NewDb(db, "pgx")}, nil
}

func NewClientWithDB(db *sql.DB) *Client {
	return &Client{db: sqlx.NewDb(db, "pgx")}
}

func initMigration(db *sql.DB) (*migrate.Migrate, error) {
	iofsDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, err
	}

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{StatementTimeout: 5 * time.Minute})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance("iofs", iofsDriver, "pgx4", driver)
}

func checkPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%w [%s]", errDuplicateKey, pgErr.Detail)
		case pgerrcode.CheckViolation:
			return fmt.Errorf("%w [%s]", errCheckViolation, pgErr.Detail)
		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%w [%s]", errForeignKeyViolation, pgErr.Detail)
		}
	}
	return err
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
