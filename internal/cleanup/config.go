package cleanup

import "time"

type Config struct {
	DryRun     bool          `mapstructure:"dry_run" default:"true"`
	ExpiryTime time.Duration `mapstructure:"expiry_time" default:"720h"`
	Services   string        `mapstructure:"services"` // list of services separated by comma, "all" means all service
}
