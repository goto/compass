package asset

import (
	"errors"
	"time"
)

var errDeleteAssetsTimeoutIsZero = errors.New("delete assets timeout must greater than 0 second")

type Config struct {
	AdditionalTypes               []string      `mapstructure:"additional_types"`
	DeleteAssetsTimeout           time.Duration `mapstructure:"delete_assets_timeout" default:"5m"`
	ExcludedChangelogPaths        []string      `mapstructure:"excluded_changelog_paths"`
	ColumnLineageHost             string        `mapstructure:"column_lineage_host"`
	ColumnLineageChangeIdentifier string        `mapstructure:"column_lineage_change_identifier"`
}

func (c *Config) Validate() error {
	if c.DeleteAssetsTimeout == 0 {
		return errDeleteAssetsTimeoutIsZero
	}

	return nil
}
