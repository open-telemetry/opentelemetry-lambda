package io.opentelemetry.lambda.sampleapps.awssdk;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;
import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.common.AttributeKey;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.metrics.LongUpDownCounter;
import io.opentelemetry.api.metrics.Meter;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.ListBucketsResponse;

public class AwsSdkRequestHandler
    implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {

  private static final Logger logger = LogManager.getLogger(AwsSdkRequestHandler.class);
  private static final Meter sampleMeter =
      GlobalOpenTelemetry.getMeterProvider()
          .meterBuilder("aws-otel")
          .setInstrumentationVersion("1.0")
          .build();
  private static final LongUpDownCounter queueSizeCounter =
      sampleMeter
          .upDownCounterBuilder("queueSizeChange")
          .setDescription("Queue Size change")
          .setUnit("one")
          .build();

  private static final AttributeKey<String> API_NAME = AttributeKey.stringKey("apiName");
  private static final AttributeKey<String> STATUS_CODE = AttributeKey.stringKey("statuscode");
  private static final Attributes METRIC_ATTRIBUTES =
      Attributes.builder().put(API_NAME, "apiName").put(STATUS_CODE, "200").build();

  @Override
  public APIGatewayProxyResponseEvent handleRequest(
      APIGatewayProxyRequestEvent input, Context context) {
    logger.info("Serving lambda request.");

    APIGatewayProxyResponseEvent response = new APIGatewayProxyResponseEvent();
    try (S3Client s3 = S3Client.create()) {
      ListBucketsResponse listBucketsResponse = s3.listBuckets();
      response.setBody(
          "Hello lambda - found " + listBucketsResponse.buckets().size() + " buckets.");
    }

    // Generate a sample counter metric using the OpenTelemetry Java Metrics API
    queueSizeCounter.add(2, METRIC_ATTRIBUTES);

    return response;
  }
}
