package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
)

// Version variable will be replaced at link time after `make` has been run.
var Version = "latest"

// GitHash variable will be replaced at link time after `make` has been run.
var GitHash = "<NOT PROPERLY GENERATED>"

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
	configFile = getConfig()
)

func getConfig() string {
	val, ex := os.LookupEnv("OPENTELEMETRY_COLLECTOR_CONFIG_FILE")
	if !ex {
		return "/opt/collector-config/config.yaml"
	}
	return val
}

// NewInProcessCollector creates a new InProcessCollector using the supplied component factories.
func NewInProcessCollector(factories component.Factories) *InProcessCollector {
	return &InProcessCollector{
		factories: factories,
	}
}

func envLoaderConfigFactory(v *viper.Viper, factories component.Factories) (*configmodels.Config, error) {
	if configContent, ok := os.LookupEnv("OPENTELEMETRY_COLLECTOR_CONFIG_CONTENT"); ok {
		println("Reading config from environment: ", configContent)
		configContent = strings.Replace(configContent, "\\n", "\n", -1)
		var configBytes = []byte(configContent)
		err := v.ReadConfig(bytes.NewBuffer(configBytes))
		if err != nil {
			return nil, fmt.Errorf("error loading config %v", err)
		}
		return config.Load(v, factories)
	}

	println("Reading config from file: ", configFile)
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
	cfg, err := envLoaderConfigFactory(v, ipp.factories)
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
			Version:  Version,
			GitHash:  GitHash,
		},
		ConfigFactory: func(v *viper.Viper, cmd *cobra.Command, factories component.Factories) (*configmodels.Config, error) {
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
		ipp.svc.Shutdown()
	}
	<-ipp.appDone
	stopped = ipp.stopped
	return stopped, err
}
