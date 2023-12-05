package job

import (
	"context"
	"time"
)

func NewService(jobRepository Repository) *Service {
	return &Service{
		jobRepository: jobRepository,
	}
}

type Service struct {
	jobRepository Repository
}

// GetSyncJobsByService handles business process to get tags by its asset id
func (s *Service) GetSyncJobsByService(ctx context.Context, serviceName string) ([]JobsQueue, error) {
	return s.jobRepository.GetSyncJobsByService(ctx, serviceName)
}

// GetSyncJobsByService handles business process to get tags by its asset id
func (s *Service) Insert(ctx context.Context, jobType string, payload []byte, runAt time.Time) (jobID string, err error) {
	return s.jobRepository.Insert(ctx, jobType, payload, runAt)
}

// GetSyncJobsByService handles business process to get tags by its asset id
func (s *Service) Delete(ctx context.Context, jobID string) error {
	return s.jobRepository.Delete(ctx, jobID)
}
