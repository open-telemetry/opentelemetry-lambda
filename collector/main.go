package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/open-telemetry/opentelemetry-lambda/collector/extension"
	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents"
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	logPrefix       = fmt.Sprintf("[%s]", extensionName)
)

func main() {
	log("Launching Opentelemetry Lambda extension, version: ", Version)
	factories, _ := lambdacomponents.Components()
	collector := NewInProcessCollector(factories)
	collector.prepareConfig()
	collector.start()

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		log("Received", s)
		log("Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	log("Register response:", prettyPrint(res))

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx, collector)
}

func processEvents(ctx context.Context, collector *InProcessCollector) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				log("Error:", err)
				log("Exiting")
				return
			}

			log("Received event:", prettyPrint(res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				collector.stop() // TODO: handle return values
				log("Received SHUTDOWN event")
				log("Exiting")
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

// log is similar to fmt.Println but it logs a
// log prefix before the log message.
func log(a ...interface{}) {
	fmt.Println(logPrefix, a)
}
