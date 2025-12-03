package user

import (
	"context"

	"github.com/goto/salt/log"
)

// Service is a type of service that manages business process
type Service struct {
	repository Repository
	logger     log.Logger
}

// ValidateUser checks if user email is already in DB
// if exist in DB, return user ID, if not exist in DB, create a new one
func (s *Service) ValidateUser(ctx context.Context, email string) (string, error) {
	if email == "" {
		return "", ErrNoUserInformation
	}

	userID, err := s.repository.GetOrInsertByEmail(ctx, &User{Email: email})
	if err != nil {
		s.logger.Error("error when GetOrInsertByEmail in ValidateUser service", "err", err.Error())
		return "", err
	}

	return userID, nil
}

// NewService initializes user service
func NewService(logger log.Logger, repository Repository, opts ...func(*Service)) *Service {
	s := &Service{
		repository: repository,
		logger:     logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}
