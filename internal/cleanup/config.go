package cleanup

import "time"

type Config struct {
	DryRun     bool          `mapstructure:"dry_run" default:"true"`
	ExpiryTime time.Duration `mapstructure:"expiry_time" default:"30d"`
	Services   string        `mapstructure:"services" default:"all"` // list of services separated by comma, "all" means all service
}
