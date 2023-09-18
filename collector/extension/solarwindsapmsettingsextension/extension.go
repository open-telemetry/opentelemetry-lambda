package solarwindsapmsettingsextension

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/solarwindscloud/apm-proto/go/collectorpb"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"
	"math"
	"os"
	"time"
)

const (
	RawOutputFile  = "/tmp/solarwinds-apm-settings-raw"
	JSONOutputFile = "/tmp/solarwinds-apm-settings.json"
)

type solarwindsapmSettingsExtension struct {
	logger *zap.Logger
	config *Config
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client collectorpb.TraceCollectorClient
}

func (extension *solarwindsapmSettingsExtension) Start(ctx context.Context, host component.Host) error {
	extension.logger.Debug("Starting up solarwinds apm settings extension")
	ctx = context.Background()
	ctx, extension.cancel = context.WithCancel(ctx)

	var err error
	extension.conn, err = grpc.Dial(extension.config.Endpoint, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return fmt.Errorf("Failed to dial: " + err.Error())
	} else {
		extension.logger.Info("Dailed to " + extension.config.Endpoint)
	}
	extension.client = collectorpb.NewTraceCollectorClient(extension.conn)

	var interval time.Duration
	interval, err = time.ParseDuration(extension.config.Interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if hostname, err := os.Hostname(); err != nil {
					extension.logger.Fatal("Unable to call os.Hostname() " + err.Error())
				} else {
					request := &collectorpb.SettingsRequest{
						ApiKey: extension.config.Key,
						Identity: &collectorpb.HostID{
							Hostname: hostname,
						},
						ClientVersion: "2",
					}
					if response, err := extension.client.GetSettings(ctx, request); err != nil {
						extension.logger.Fatal("Unable to getSettings from " + extension.config.Endpoint + " " + err.Error())
					} else {
						switch result := response.GetResult(); result {
						case collectorpb.ResultCode_OK:
							if bytes, err := proto.Marshal(response); err != nil {
								extension.logger.Error("Unable to marshal response to bytes " + err.Error())
							} else {
								// Output in raw format
								if err := os.WriteFile(RawOutputFile, bytes, 0644); err != nil {
									extension.logger.Error("Unable to write " + RawOutputFile + " " + err.Error())
								} else {
									extension.logger.Info(RawOutputFile + " is refreshed")
								}
							}
							// Output in human-readable format
							var settings []map[string]interface{}
							for _, item := range response.GetSettings() {

								marshalOptions := protojson.MarshalOptions{
									EmitUnpopulated: true,
								}
								if settingBytes, err := marshalOptions.Marshal(item); err != nil {
									extension.logger.Warn("Error to marshal setting JSON[] byte from response.GetSettings() " + err.Error())
								} else {
									setting := make(map[string]interface{})
									if err := json.Unmarshal(settingBytes, &setting); err != nil {
										extension.logger.Warn("Error to unmarshal setting JSON object from setting JSON[]byte " + err.Error())
									} else {
										if _, ok := setting["flags"]; ok {
											setting["flags"] = string(item.Flags)
										}
										if arguments, ok := setting["arguments"].(map[string]interface{}); ok {
											if value, ok := item.Arguments["BucketCapacity"]; ok {
												arguments["BucketCapacity"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
											}
											if value, ok := item.Arguments["BucketRate"]; ok {
												arguments["BucketRate"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
											}
											if value, ok := item.Arguments["TriggerRelaxedBucketCapacity"]; ok {
												arguments["TriggerRelaxedBucketCapacity"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
											}
											if value, ok := item.Arguments["TriggerRelaxedBucketRate"]; ok {
												arguments["TriggerRelaxedBucketRate"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
											}
											if value, ok := item.Arguments["TriggerStrictBucketCapacity"]; ok {
												arguments["TriggerStrictBucketCapacity"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
											}
											if value, ok := item.Arguments["TriggerStrictBucketRate"]; ok {
												arguments["TriggerStrictBucketRate"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
											}

											if value, ok := item.Arguments["MetricsFlushInterval"]; ok {
												arguments["MetricsFlushInterval"] = int32(binary.LittleEndian.Uint32(value))
											}
											if value, ok := item.Arguments["MaxTransactions"]; ok {
												arguments["MaxTransactions"] = int32(binary.LittleEndian.Uint32(value))
											}
											if value, ok := item.Arguments["MaxCustomMetrics"]; ok {
												arguments["MaxCustomMetrics"] = int32(binary.LittleEndian.Uint32(value))
											}
											if value, ok := item.Arguments["EventsFlushInterval"]; ok {
												arguments["EventsFlushInterval"] = int32(binary.LittleEndian.Uint32(value))
											}
											if value, ok := item.Arguments["ProfilingInterval"]; ok {
												arguments["ProfilingInterval"] = int32(binary.LittleEndian.Uint32(value))
											}

											if value, ok := item.Arguments["SignatureKey"]; ok {
												arguments["SignatureKey"] = string(value)
											}
										}
										settings = append(settings, setting)
									}
								}
							}
							if content, err := json.Marshal(settings); err != nil {
								extension.logger.Warn("Error to marshal setting JSON[] byte from settings " + err.Error())
							} else {
								if err := os.WriteFile(JSONOutputFile, content, 0644); err != nil {
									extension.logger.Error("Unable to write " + JSONOutputFile + " " + err.Error())
								} else {
									extension.logger.Info(JSONOutputFile + " is refreshed")
								}
							}
						case collectorpb.ResultCode_TRY_LATER:
							extension.logger.Warn("GetSettings returned TRY_LATER " + response.GetWarning())
						case collectorpb.ResultCode_INVALID_API_KEY:
							extension.logger.Warn("GetSettings returned INVALID_API_KEY " + response.GetWarning())
						case collectorpb.ResultCode_LIMIT_EXCEEDED:
							extension.logger.Warn("GetSettings returned LIMIT_EXCEEDED " + response.GetWarning())
						case collectorpb.ResultCode_REDIRECT:
							extension.logger.Warn("GetSettings returned REDIRECT " + response.GetWarning())
						default:
							extension.logger.Warn("Unknown ResultCode from GetSettings " + response.GetWarning())
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (extension *solarwindsapmSettingsExtension) Shutdown(ctx context.Context) error {
	extension.logger.Debug("Shutting down solarwinds apm settings extension")
	extension.conn.Close()
	extension.cancel()
	return nil
}
