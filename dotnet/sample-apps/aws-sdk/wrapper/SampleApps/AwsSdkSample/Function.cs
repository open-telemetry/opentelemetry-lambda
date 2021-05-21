using Amazon.Lambda.APIGatewayEvents;
using Amazon.Lambda.Core;
using Amazon.S3;
using OpenTelemetry;
using OpenTelemetry.Contrib.Instrumentation.AWSLambda.Implementation;
using OpenTelemetry.Trace;
using System;

// Assembly attribute to enable the Lambda function's JSON input to be converted into a .NET class.
[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.Json.JsonSerializer))]

namespace AwsSdkSample
{
    public class Function
    {
        public static TracerProvider tracerProvider;

        static Function()
        {
            AppContext.SetSwitch("System.Net.Http.SocketsHttpHandler.Http2UnencryptedSupport", true);

            tracerProvider = Sdk.CreateTracerProviderBuilder()
                .AddAWSInstrumentation()
                .AddOtlpExporter()
                .AddAWSLambdaConfigurations()
                .Build();
        }

        // use AwsSdkSample::AwsSdkSample.Function::TracingFunctionHandler as input Lambda handler instead
        public APIGatewayProxyResponse TracingFunctionHandler(APIGatewayProxyRequest request, ILambdaContext context)
        {
            return AWSLambdaWrapper.Trace(tracerProvider, FunctionHandler, request, context);
        }

        /// <summary>
        /// A simple function that takes a APIGatewayProxyRequest and returns a APIGatewayProxyResponse
        /// </summary>
        /// <param name="input"></param>
        /// <param name="context"></param>
        /// <returns></returns>
        public APIGatewayProxyResponse FunctionHandler(APIGatewayProxyRequest request, ILambdaContext context)
        {
            var S3Client = new AmazonS3Client();
            _ = S3Client.ListBucketsAsync().Result;
            return new APIGatewayProxyResponse() { StatusCode = 200, Body = "Hello Validator!" };
        }
    }
}
