package postgres

import (
	"context"
	"fmt"

	"github.com/goto/compass/core/job"
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
