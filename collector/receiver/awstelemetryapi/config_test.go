package awstelemetryapi

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	factory := NewFactory("test-extension-id")
	defaultCfg := factory.CreateDefaultConfig()

	// Test that the default config is created correctly
	require.Equal(t, uint(defaultMaxItems), defaultCfg.(*Config).MaxItems)

	// Test loading a config from YAML
	sub, err := cm.Sub(component.NewIDWithName(component.MustNewType(typeStr), "2").String())
	require.NoError(t, err)

	cfg := factory.CreateDefaultConfig()
	require.NoError(t, sub.Unmarshal(cfg))

	expected := &Config{
		extensionID: "test-extension-id",
		Port:        12345,
		Types:       []string{platform},
		MaxItems:    defaultMaxItems,  // Should still be default
		MaxBytes:    defaultMaxBytes,  // Should still be default
		TimeoutMS:   defaultTimeoutMS, // Should still be default
	}
	require.Equal(t, expected, cfg)
}
