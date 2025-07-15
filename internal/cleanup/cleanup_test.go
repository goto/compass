package cleanup_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/goto/compass/internal/cleanup"
	"github.com/goto/compass/internal/server/v1beta1/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	cfg := cleanup.Config{
		DryRun:         true,
		Services:       "svc1,svc2",
		ExpiryDuration: 24 * time.Hour,
	}

	tests := []struct {
		name        string
		mockSetup   func(*mocks.AssetService)
		expectCount uint32
		expectErr   string
	}{
		{
			name: "success",
			mockSetup: func(mockSvc *mocks.AssetService) {
				mockSvc.On("DeleteAssetsByServicesAndUpdatedAt", mock.Anything, cfg.DryRun, cfg.Services, cfg.ExpiryDuration).Return(uint32(5), nil)
			},
			expectCount: 5,
			expectErr:   "",
		},
		{
			name: "error from service",
			mockSetup: func(mockSvc *mocks.AssetService) {
				errExpected := errors.New("service error")
				mockSvc.On("DeleteAssetsByServicesAndUpdatedAt", mock.Anything, cfg.DryRun, cfg.Services, cfg.ExpiryDuration).Return(uint32(0), errExpected)
			},
			expectCount: 0,
			expectErr:   "failed to cleanup assets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mocks.AssetService)
			tt.mockSetup(mockSvc)

			count, err := cleanup.Run(ctx, cfg, mockSvc)
			if tt.expectErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
			}
			assert.Equal(t, tt.expectCount, count)
			mockSvc.AssertExpectations(t)
		})
	}
}
