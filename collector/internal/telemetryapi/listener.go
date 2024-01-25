// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetryapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"go.uber.org/zap"
)

const defaultListenerPort = "53612"
const initialQueueSize = 5

// Listener is used to listen to the Telemetry API
type Listener struct {
	httpServer *http.Server
	logger     *zap.Logger
	// queue is a synchronous queue and is used to put the received log events to be dispatched later
	queue *queue.Queue
}

func NewListener(logger *zap.Logger) *Listener {
	return &Listener{
		httpServer: nil,
		logger:     logger.Named("telemetryAPI.Listener"),
		queue:      queue.New(initialQueueSize),
	}
}

func listenOnAddress() string {
	envAwsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && envAwsLocal == "true" {
		addr = ":" + defaultListenerPort
	} else {
		addr = "sandbox.localdomain:" + defaultListenerPort
	}

	return addr
}

// Start the server in a goroutine where the log events will be sent
func (s *Listener) Start() (string, error) {
	address := listenOnAddress()
	s.logger.Info("Listening for requests", zap.String("address", address))
	s.httpServer = &http.Server{Addr: address}
	http.HandleFunc("/", s.httpHandler)
	go func() {
		err := s.httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			s.logger.Error("Unexpected stop on HTTP Server", zap.Error(err))
			s.Shutdown()
		} else {
			s.logger.Info("HTTP Server closed:", zap.Error(err))
		}
	}()
	return fmt.Sprintf("http://%s/", address), nil
}

// httpHandler handles the requests coming from the Telemetry API.
// Everytime Telemetry API sends log events, this function will read them from the response body
// and put into a synchronous queue to be dispatched later.
// Logging or printing besides the error cases below is not recommended if you have subscribed to
// receive extension logs. Otherwise, logging here will cause Telemetry API to send new logs for
// the printed lines which may create an infinite loop.
func (s *Listener) httpHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("error reading body", zap.Error(err))
		return
	}

	// Parse and put the log messages into the queue
	var slice []Event
	_ = json.Unmarshal(body, &slice)

	for _, el := range slice {
		if err := s.queue.Put(el); err != nil {
			s.logger.Error("Failed to put event in queue", zap.Error(err))
		}
	}

	s.logger.Debug("logEvents received", zap.Int("count", len(slice)), zap.Int64("queue_length", s.queue.Len()))
	slice = nil
}

// Shutdown the HTTP server listening for logs
func (s *Listener) Shutdown() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			s.logger.Error("Failed to shutdown HTTP server gracefully", zap.Error(err))
		} else {
			s.httpServer = nil
		}
	}
}

func (s *Listener) Wait(ctx context.Context, reqID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.logger.Debug("looking for platform.runtimeDone event")
			items, err := s.queue.Get(10)
			if err != nil {
				return fmt.Errorf("unable to get telemetry events from queue: %w", err)
			}

			for _, item := range items {
				i, ok := item.(Event)
				if !ok {
					s.logger.Warn("non-Event found in queue", zap.Any("item", item))
					continue
				}
				s.logger.Debug("Event processed", zap.Any("event", i))
				if i.Type != "platform.runtimeDone" {
					continue
				}

				if i.Record["requestId"] == reqID {
					return nil
				}
			}
		}
	}
}
