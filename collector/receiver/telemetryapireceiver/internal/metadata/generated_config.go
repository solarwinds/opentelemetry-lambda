// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/confmap"
)

// MetricConfig provides common config for a particular metric.
type MetricConfig struct {
	Enabled bool `mapstructure:"enabled"`

	enabledSetByUser bool
}

func (ms *MetricConfig) Unmarshal(parser *confmap.Conf) error {
	if parser == nil {
		return nil
	}
	err := parser.Unmarshal(ms)
	if err != nil {
		return err
	}
	ms.enabledSetByUser = parser.IsSet("enabled")
	return nil
}

// MetricsConfig provides config for telemetryapi metrics.
type MetricsConfig struct {
	FaasColdstarts  MetricConfig `mapstructure:"faas.coldstarts"`
	FaasErrors      MetricConfig `mapstructure:"faas.errors"`
	FaasInvocations MetricConfig `mapstructure:"faas.invocations"`
	FaasTimeouts    MetricConfig `mapstructure:"faas.timeouts"`
}

func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		FaasColdstarts: MetricConfig{
			Enabled: true,
		},
		FaasErrors: MetricConfig{
			Enabled: true,
		},
		FaasInvocations: MetricConfig{
			Enabled: true,
		},
		FaasTimeouts: MetricConfig{
			Enabled: true,
		},
	}
}

// MetricsBuilderConfig is a configuration for telemetryapi metrics builder.
type MetricsBuilderConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
}

func DefaultMetricsBuilderConfig() MetricsBuilderConfig {
	return MetricsBuilderConfig{
		Metrics: DefaultMetricsConfig(),
	}
}
