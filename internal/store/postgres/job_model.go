package postgres

import (
	"database/sql"
	"time"

	"github.com/goto/compass/core/job"
)

// JobQueueModel is a model for tag value in database table
type JobQueueModel struct {
	ID            string         `db:"id"`
	Type          string         `db:"type"`
	LastError     sql.NullString `db:"last_error"`
	AttemptsDone  int32          `db:"attempts_done"`
	Payload       []byte         `db:"payload"`
	RunAt         time.Time      `db:"run_at"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	LastAttemptAt sql.NullTime   `db:"last_attempt_at"`
}

type JobQueueModels []JobQueueModel

func (j *JobQueueModel) toJobQueue() job.JobsQueue {
	return job.JobsQueue{
		ID:            j.ID,
		Type:          j.Type,
		LastError:     j.LastError.String,
		AttemptsDone:  j.AttemptsDone,
		Payload:       j.Payload,
		RunAt:         j.RunAt,
		CreatedAt:     j.CreatedAt,
		UpdatedAt:     j.UpdatedAt,
		LastAttemptAt: j.LastAttemptAt.Time,
	}
}

func (j JobQueueModels) toJobQueues() []job.JobsQueue {
	if len(j) == 0 {
		return nil
	}

	jobsQueues := []job.JobsQueue{}
	for _, v := range j {
		jobsQueues = append(jobsQueues, v.toJobQueue())
	}
	return jobsQueues
}
