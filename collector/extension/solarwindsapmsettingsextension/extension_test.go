package solarwindsapmsettingsextension

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/extension"
)

func TestCreateExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "default",
			cfg: &Config{
				ClientConfig: configgrpc.ClientConfig{
					Endpoint: DefaultEndpoint,
				},
				Interval: DefaultInterval,
			},
		},
		{
			name: "anything",
			cfg: &Config{
				ClientConfig: configgrpc.ClientConfig{
					Endpoint: "0.0.0.0:1234",
				},
				Key:      "something",
				Interval: time.Duration(10) * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex := createAnExtension(tt.cfg, t)
			require.NoError(t, ex.Shutdown(context.TODO()))
		})
	}
}

// NewNopSettings returns a new nop settings for extension.Factory Create* functions.
func newNopSettings() extension.Settings {
	return extension.Settings{
		ID:                component.NewIDWithName(component.MustNewType("nop"), uuid.NewString()),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
}

// create extension
func createAnExtension(c *Config, t *testing.T) extension.Extension {
	ex, err := newSolarwindsApmSettingsExtension(c, NewNopSettings())
	require.NoError(t, err)
	err = ex.Start(context.TODO(), nil)
	require.NoError(t, err)
	return ex
}
