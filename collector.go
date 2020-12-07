package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"
)

// InProcessCollector implements the OtelcolRunner interfaces running a single otelcol as a go routine within the
// same process as the test executor.
type InProcessCollector struct {
	factories component.Factories
	config    *configmodels.Config
	svc       *service.Application
	appDone   chan struct{}
	stopped   bool
}

var (
	configFile = os.Getenv("OPENTELEMETRY_COLLECTOR_CONFIG_FILE")
)

// NewInProcessCollector creates a new InProcessCollector using the supplied component factories.
func NewInProcessCollector(factories component.Factories) *InProcessCollector {
	return &InProcessCollector{
		factories: factories,
	}
}

// envFileLoaderConfigFactory implements ConfigFactory and it creates configuration from file.
func envFileLoaderConfigFactory(v *viper.Viper, factories component.Factories) (*configmodels.Config, error) {
	println("Loading config file:", configFile)
	file := configFile
	if file == "" {
		return nil, errors.New("config file not specified")
	}
	v.SetConfigFile(file)
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config file %q: %v", file, err)
	}
	return config.Load(v, factories)
}

func (ipp *InProcessCollector) prepareConfig() (err error) {
	v := config.NewViper()
	v.SetConfigType("yaml")
	cfg, err := envFileLoaderConfigFactory(v, ipp.factories)
	if err != nil {
		return err
	}
	err = config.ValidateConfig(cfg, zap.NewNop())
	if err != nil {
		return err
	}
	ipp.config = cfg
	return err
}

func (ipp *InProcessCollector) start() error {
	params := service.Parameters{
		ApplicationStartInfo: component.ApplicationStartInfo{
			ExeName:  "otelcol",
			LongName: "InProcess Collector",
			// TODO: set versions
			// Version:  version.Version,
			// GitHash:  version.GitHash,
		},
		ConfigFactory: func(v *viper.Viper, factories component.Factories) (*configmodels.Config, error) {
			return ipp.config, nil
		},
		Factories: ipp.factories,
	}
	var err error
	ipp.svc, err = service.New(params)
	if err != nil {
		return err
	}
	ipp.svc.Command().SetArgs([]string{"--metrics-level=NONE"})

	ipp.appDone = make(chan struct{})
	go func() {
		defer close(ipp.appDone)
		appErr := ipp.svc.Run()
		if appErr != nil {
			err = appErr
		}
	}()

	for state := range ipp.svc.GetStateChannel() {
		switch state {
		case service.Starting:
			// NoOp
		case service.Running:
			return err
		default:
			err = fmt.Errorf("unable to start, otelcol state is %d", state)
		}
	}
	return err
}

func (ipp *InProcessCollector) stop() (stopped bool, err error) {
	if !ipp.stopped {
		ipp.stopped = true
		ipp.svc.SignalTestComplete()
	}
	<-ipp.appDone
	stopped = ipp.stopped
	return stopped, err
}
