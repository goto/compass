package job

import "context"

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
