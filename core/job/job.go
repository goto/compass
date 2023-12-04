package job

import (
	"context"
	"time"
)

type Repository interface {
	GetSyncJobsByService(ctx context.Context, serviceName string) ([]JobsQueue, error)
}

type JobsQueue struct {
	ID          string    `db:"id"`
	Type        string    `db:"type"`
	LastError   string    `db:"last_error"`
	AttemptsDo  int32     `db:"attempts_do"`
	Payload     []byte    `db:"payload"`
	RunAt       time.Time `db:"run_at"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	LastAttempt time.Time `db:"last_attempt"`
}
