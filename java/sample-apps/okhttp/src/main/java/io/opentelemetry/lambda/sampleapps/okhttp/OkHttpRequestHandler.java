package io.opentelemetry.lambda.sampleapps.okhttp;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;
import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.instrumentation.okhttp.v3_0.OkHttpTelemetry;
import java.io.IOException;
import java.io.UncheckedIOException;
import okhttp3.Call;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

public class OkHttpRequestHandler
    implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {

  private static final Logger logger = LogManager.getLogger(OkHttpRequestHandler.class);

  @Override
  public APIGatewayProxyResponseEvent handleRequest(
      APIGatewayProxyRequestEvent input, Context context) {
    logger.info("Serving lambda request.");

    OkHttpClient baseClient = new OkHttpClient();
    Call.Factory callFactory =
        OkHttpTelemetry.create(GlobalOpenTelemetry.get()).newCallFactory(baseClient);

    APIGatewayProxyResponseEvent response = new APIGatewayProxyResponseEvent();

    Request request = new Request.Builder().url("https://aws.amazon.com/").build();
    try (Response okhttpResponse = callFactory.newCall(request).execute()) {
      response.setBody(
          "Hello lambda - fetched " + okhttpResponse.body().string().length() + " bytes.");
    } catch (IOException e) {
      throw new UncheckedIOException("Could not fetch with okhttp", e);
    }
    return response;
  }
}
