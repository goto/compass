package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/goto/compass/core/job"
	"github.com/oklog/ulid/v2"
)

// JobRepository is a type that manages jobs queue operation ot the primary database
type JobRepository struct {
	client *Client
}

func (r *JobRepository) GetSyncJobsByService(ctx context.Context, service string) ([]job.JobsQueue, error) {
	var res JobQueueModels
	query := `SELECT * FROM jobs_queue WHERE payload::text = $1 AND type = 'sync-asset'`
	if err := r.client.db.SelectContext(ctx, &res, query, service); err != nil {
		return nil, fmt.Errorf("get sync jobs by service: %w", err)
	}

	return res.toJobQueues(), nil
}

func (r *JobRepository) Insert(ctx context.Context, jobType string, payload []byte, runAt time.Time) (string, error) {
	var jobID string
	query := `
		INSERT INTO
			jobs_queue
			(id, type, payload, run_at)
		VALUES
			($1, $2, $3, $4)
		RETURNING id
		`
	if err := r.client.db.QueryRowContext(ctx, query, ulid.Make().String(), jobType, payload, runAt).Scan(&jobID); err != nil {
		return "", fmt.Errorf("insert job queue: %w", err)
	}

	return jobID, nil
}

func (r *JobRepository) Delete(ctx context.Context, jobID string) error {
	query := `DELETE FROM jobs_queue WHERE id = $1`
	if _, err := r.client.db.ExecContext(ctx, query, jobID); err != nil {
		return fmt.Errorf("delete job queue: %w", err)
	}

	return nil
}

// NewJobRepository initializes jobs queue repository
// all methods in jobs queue repository uses passed by reference
// which will mutate the reference variable in method's argument
func NewJobRepository(client *Client) (*JobRepository, error) {
	if client == nil {
		return nil, errNilPostgresClient
	}
	return &JobRepository{
		client: client,
	}, nil
}
