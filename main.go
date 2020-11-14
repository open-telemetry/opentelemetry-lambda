package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/lightstep/opentelemetry-lambda-extension/extension"
	"github.com/spf13/viper"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
	"go.opentelemetry.io/collector/service/defaultcomponents"
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	configFile      = os.Getenv("OPENTELEMETRY_COLLECTOR_CONFIG_FILE")
	printPrefix     = fmt.Sprintf("[%s]", extensionName)
)

// envFileLoaderConfigFactory implements ConfigFactory and it creates configuration from file.
func envFileLoaderConfigFactory(v *viper.Viper, factories component.Factories) (*configmodels.Config, error) {
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

func startCollector() {
	factories, err := defaultcomponents.Components()
	if err != nil {
		log.Fatalf("failed to build default components: %v", err)
	}

	info := component.ApplicationStartInfo{
		ExeName:  "otelcol",
		LongName: "OpenTelemetry Collector",
		// Version:  version.Version,
		// GitHash:  version.GitHash,
	}

	if err := run(service.Parameters{ApplicationStartInfo: info, Factories: factories, ConfigFactory: envFileLoaderConfigFactory}); err != nil {
		log.Fatal(err)
	}
}

func run(params service.Parameters) error {
	return runInteractive(params)
}

func runInteractive(params service.Parameters) error {
	app, err := service.New(params)
	if err != nil {
		return fmt.Errorf("failed to construct the application: %w", err)
	}

	err = app.Run()
	if err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}
	return err
}

func main() {
	go startCollector()
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(printPrefix, "Received", s)
		println(printPrefix, "Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	println(printPrefix, "Register response:", prettyPrint(res))

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx)
}

func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			println(printPrefix, "Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				println(printPrefix, "Error:", err)
				println(printPrefix, "Exiting")
				return
			}
			println(printPrefix, "Received event:", prettyPrint(res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				println(printPrefix, "Received SHUTDOWN event")
				println(printPrefix, "Exiting")
				return
			}
		}
	}
}

func prettyPrint(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(data)
}
