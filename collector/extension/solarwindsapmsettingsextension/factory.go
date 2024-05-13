package solarwindsapmsettingsextension

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"

	"github.com/open-telemetry/opentelemetry-lambda/collector/extension/solarwindsapmsettingsextension/internal/metadata"
)

func createExtension(_ context.Context, settings extension.CreateSettings, cfg component.Config) (extension.Extension, error) {
	return newSolarwindsApmSettingsExtension(cfg.(*Config), settings.Logger)
}

func NewFactory() extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		createDefaultConfig,
		createExtension,
		metadata.ExtensionStability,
	)
}
