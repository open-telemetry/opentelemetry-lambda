// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// This is a minimal fork of github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor
// Synced from: v0.145.0
// Only the env and lambda detectors are included.
//
// To sync with a new upstream version:
//  1. Copy factory.go, config.go, resourcedetection_processor.go from the new version
//  2. Copy internal/resourcedetection.go, internal/context.go, internal/env/, internal/aws/lambda/ from the new version
//  3. Copy internal/metadata/generated_status.go and internal/metadata/generated_feature_gates.go from the new version
//  4. Re-apply the factory.go + config.go trimming (keep only env + lambda)
//  5. Update go.mod versions to match the new collector-contrib release
//  6. Run go mod tidy in this directory and in collector/lambdacomponents/

package resourcedetectionprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"

import (
	"time"

	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor/internal"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor/internal/aws/lambda"
)

// Config defines configuration for Resource processor.
type Config struct {
	// Detectors is an ordered list of named detectors that should be
	// run to attempt to detect resource information.
	Detectors []string `mapstructure:"detectors"`
	// Override indicates whether any existing resource attributes
	// should be overridden or preserved. Defaults to true.
	Override bool `mapstructure:"override"`
	// DetectorConfig is a list of settings specific to all detectors
	DetectorConfig DetectorConfig `mapstructure:",squash"`
	// HTTP client settings for the detector
	// Timeout default is 5s
	confighttp.ClientConfig `mapstructure:",squash"`
	// If > 0, periodically re-run detection for all configured detectors.
	// When 0 (default), no periodic refresh occurs.
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
}

// DetectorConfig contains user-specified configurations unique to all individual detectors
type DetectorConfig struct {
	// LambdaConfig contains user-specified configurations for the lambda detector
	LambdaConfig lambda.Config `mapstructure:"lambda"`
}

func detectorCreateDefaultConfig() DetectorConfig {
	return DetectorConfig{
		LambdaConfig: lambda.CreateDefaultConfig(),
	}
}

func (d *DetectorConfig) GetConfigFromType(detectorType internal.DetectorType) internal.DetectorConfig {
	switch detectorType {
	case lambda.TypeStr:
		return d.LambdaConfig
	default:
		return nil
	}
}
