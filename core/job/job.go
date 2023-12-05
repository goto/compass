package job

import (
	"context"
	"time"
)

type Repository interface {
	GetSyncJobsByService(ctx context.Context, serviceName string) ([]JobsQueue, error)
	Insert(ctx context.Context, jobType string, payload []byte, runAt time.Time) (jobID string, err error)
	Delete(ctx context.Context, jobID string) error
}

type JobsQueue struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	LastError     string    `json:"last_error,omitempty"`
	AttemptsDone  int32     `json:"attempts_done"`
	Payload       []byte    `json:"payload"`
	RunAt         time.Time `json:"run_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastAttemptAt time.Time `json:"last_attempt_at,omitempty"`
}
