package swotracemetricsconnector

import (
	"fmt"
	"time"
)

type Config struct {
	FlushInterval string `mapstructure:"flush_interval"`
}

func (config *Config) Validate() error {
	if _, err := time.ParseDuration(config.FlushInterval); err != nil {
		return fmt.Errorf("interval has to be a duration string. Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"")
	}
	return nil
}
