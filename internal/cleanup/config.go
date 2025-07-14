package cleanup

import "time"

type Config struct {
	DryRun         bool          `mapstructure:"dry_run" default:"true"`
	ExpiryDuration time.Duration `mapstructure:"expiry_duration" default:"720h0m0s"` // 30 days
	Services       string        `mapstructure:"services"`                           // list of services separated by comma, "all" means all service
}
