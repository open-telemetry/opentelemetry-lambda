package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
)

func lambda_handler(ctx context.Context) func(ctx context.Context) (interface{}, error) {
	// init aws config
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	// instrument all aws clients
	otelaws.AppendMiddlewares(&cfg.APIOptions)

	// S3
	s3Client := s3.NewFromConfig(cfg)
	// HTTP
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
		),
	}

	return func(ctx context.Context) (interface{}, error) {
		input := &s3.ListBucketsInput{}
		result, err := s3Client.ListBuckets(ctx, input)
		if err != nil {
			fmt.Printf("Got an error retrieving buckets, %v", err)
		}

		fmt.Println("Buckets:")
		for _, bucket := range result.Buckets {
			fmt.Println(*bucket.Name + ": " + bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
		}
		fmt.Println("End Buckets.")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/open-telemetry/opentelemetry-go/releases/latest", nil)
		if err != nil {
			fmt.Printf("failed to create http request, %v\n", err)
		}
		res, err := httpClient.Do(req)
		if err != nil {
			fmt.Printf("failed to make http request, %v\n", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close http response body, %v\n", err)
			}
		}(res.Body)

		var data map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			fmt.Printf("failed to read http response body, %v\n", err)
		}
		fmt.Printf("Latest OTel Go Release is '%s'\n", data["name"])

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       os.Getenv("_X_AMZN_TRACE_ID"),
		}, nil
	}
}

func main() {
	ctx := context.Background()

	tp, err := xrayconfig.NewTracerProvider(ctx)
	if err != nil {
		fmt.Printf("error creating tracer provider: %v", err)
	}

	defer func(ctx context.Context) {
		err := tp.Shutdown(ctx)
		if err != nil {
			fmt.Printf("error shutting down tracer provider: %v", err)
		}
	}(ctx)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	lambda.Start(otellambda.InstrumentHandler(lambda_handler(ctx), xrayconfig.WithRecommendedOptions(tp)... ))
}
