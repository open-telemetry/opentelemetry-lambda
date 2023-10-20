package decoupleprocessor

import (
	"context"
	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/processor/processortest"
	"testing"
	"time"
)

type MockLifecycleNotifier struct {
	listener lambdalifecycle.Listener
}

func (m *MockLifecycleNotifier) AddListener(l lambdalifecycle.Listener) {
	m.listener = l
}

type MockConsumer struct {
	info         client.Info
	dataReceived chan any
}

func (m *MockConsumer) consume(ctx context.Context, data any) error {
	m.info = client.FromContext(ctx)
	m.dataReceived <- data
	return nil
}

func newMockConsumer() *MockConsumer {
	m := MockConsumer{
		info:         client.Info{},
		dataReceived: make(chan any),
	}
	return &m
}

func Test_newDecoupleProcessor(t *testing.T) {
	notifier := &MockLifecycleNotifier{}
	config := &Config{MaxQueueSize: 1}
	type args struct {
		cfg      *Config
		consumer decoupleConsumer
	}
	tests := []struct {
		name     string
		notifier lambdalifecycle.Notifier
		args     args
		wantErr  bool
	}{
		{
			name: "No Lifecycle notifier set",
			args: args{
				cfg:      config,
				consumer: newMockConsumer(),
			},
			notifier: nil,
			wantErr:  true,
		},
		{
			name:     "Successful creation",
			notifier: notifier,
			args: args{
				cfg:      config,
				consumer: &MockConsumer{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lambdalifecycle.SetNotifier(tt.notifier)

			dp, err := newDecoupleProcessor(tt.args.cfg, tt.args.consumer, processortest.NewNopCreateSettings())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Equal(t, nil, err)
			}

			if tt.notifier == notifier {
				require.Equal(t, notifier.listener, dp, "newDecoupleProcessor() did not register a lifecycle listener!")
			}
		})
	}
}

func TestLifecycle(t *testing.T) {
	consumer := newMockConsumer()
	notifier := &MockLifecycleNotifier{}
	config := &Config{MaxQueueSize: 1}
	lambdalifecycle.SetNotifier(notifier)

	t.Run("create and otelcol shutdown only", func(t *testing.T) {
		dp, err := newDecoupleProcessor(config, consumer, processortest.NewNopCreateSettings())
		require.NoError(t, err)
		require.NoError(t, dp.shutdown(context.Background()))
	})

	t.Run("full lifecycle", func(t *testing.T) {
		dp, err := newDecoupleProcessor(config, consumer, processortest.NewNopCreateSettings())
		require.NoError(t, err)

		dp.FunctionInvoked()
		dp.FunctionFinished()
		dp.EnvironmentShutdown()

		require.NoError(t, dp.shutdown(context.Background()))
	})

	t.Run("full lifecycle with data from function", func(t *testing.T) {
		dp, err := newDecoupleProcessor(config, consumer, processortest.NewNopCreateSettings())
		require.NoError(t, err)

		dp.FunctionInvoked()

		// Check that data waiting to be sent delays the completion of FunctionFinished()
		expectedData := "data"
		dp.queueData(client.NewContext(context.Background(), client.Info{}), expectedData)
		start := time.Now()
		var data any
		go func() {
			time.Sleep(200 * time.Millisecond)
			data = <-consumer.dataReceived
		}()
		dp.FunctionFinished()
		finish := time.Now()
		require.WithinRange(t, finish, start.Add(200*time.Millisecond), start.Add(300*time.Millisecond))
		require.Equal(t, expectedData, data)

		dp.EnvironmentShutdown()
		require.NoError(t, dp.shutdown(context.Background()))
	})

	t.Run("full lifecycle with data before shutdown", func(t *testing.T) {
		dp, err := newDecoupleProcessor(config, consumer, processortest.NewNopCreateSettings())
		require.NoError(t, err)

		dp.FunctionInvoked()
		dp.FunctionFinished()
		dp.EnvironmentShutdown()

		// Check that data waiting to be sent delays the completion of shutdown()
		expectedData := "data"
		dp.queueData(client.NewContext(context.Background(), client.Info{}), expectedData)
		start := time.Now()
		var data any
		go func() {
			time.Sleep(200 * time.Millisecond)
			data = <-consumer.dataReceived
		}()
		require.NoError(t, dp.shutdown(context.Background()))
		finish := time.Now()
		require.WithinRange(t, finish, start.Add(200*time.Millisecond), start.Add(300*time.Millisecond))
		require.Equal(t, expectedData, data)
	})
}
