package asset_test

import (
	"testing"

	"github.com/goto/compass/core/asset"
	"github.com/stretchr/testify/assert"
)

func TestLineageDirection_IsValid(t *testing.T) {
	testCases := []struct {
		Description string
		Direction   asset.LineageDirection
		Expected    bool
	}{
		{
			Description: "upstream direction is valid",
			Direction:   asset.LineageDirectionUpstream,
			Expected:    true,
		},
		{
			Description: "downstream direction is valid",
			Direction:   asset.LineageDirectionDownstream,
			Expected:    true,
		},
		{
			Description: "empty direction is valid",
			Direction:   "",
			Expected:    true,
		},
		{
			Description: "unknown direction is invalid",
			Direction:   asset.LineageDirection("unknown"),
			Expected:    false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Direction.IsValid())
		})
	}
}

func TestLineageCoverage_IsValid(t *testing.T) {
	testCases := []struct {
		Description string
		Coverage    asset.LineageCoverage
		Expected    bool
	}{
		{
			Description: "asset coverage is valid",
			Coverage:    asset.LineageCoverageAsset,
			Expected:    true,
		},
		{
			Description: "column coverage is valid",
			Coverage:    asset.LineageCoverageColumn,
			Expected:    true,
		},
		{
			Description: "empty coverage is valid",
			Coverage:    "",
			Expected:    true,
		},
		{
			Description: "unknown coverage is invalid",
			Coverage:    asset.LineageCoverage("unknown"),
			Expected:    false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Coverage.IsValid())
		})
	}
}
