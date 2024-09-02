package asset

import (
	"errors"
	"time"
)

var errDeleteAssetsTimeoutIsZero = errors.New("delete assets timeout must greater than 0 second")

type Config struct {
	AdditionalTypes     []string      `mapstructure:"additional_types"`
	DeleteAssetsTimeout time.Duration `mapstructure:"delete_assets_timeout" default:"5m"`
}

func (c *Config) Validate() error {
	if c.DeleteAssetsTimeout == 0 {
		return errDeleteAssetsTimeoutIsZero
	}

	return nil
}
