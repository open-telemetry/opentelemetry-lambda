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
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

func lambda_handler(ctx context.Context) (interface{}, error) {
	// init aws config
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	// instrument all aws clients
	otelaws.AppendMiddlewares(&cfg.APIOptions)

	// S3
	s3Client := s3.NewFromConfig(cfg)
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

	// HTTP
	orig := otelhttp.DefaultClient
	otelhttp.DefaultClient = &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		),
	}
	defer func() { otelhttp.DefaultClient = orig }()
	res, err := otelhttp.Get(ctx, "https://api.github.com/repos/open-telemetry/opentelemetry-go/releases/latest")
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

func main() {
	lambda.Start(otellambda.LambdaHandlerWrapper(lambda_handler))
}