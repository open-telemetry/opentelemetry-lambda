package io.opentelemetry.lambda.integrationtests;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;
import software.amazon.awssdk.services.sts.StsClient;
import software.amazon.awssdk.services.sts.model.GetCallerIdentityResponse;

public class StsRequestHandler
    implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {

  @Override
  public APIGatewayProxyResponseEvent handleRequest(
      APIGatewayProxyRequestEvent event, Context context) {
    String account;
    try (StsClient sts = StsClient.create()) {
      GetCallerIdentityResponse identity = sts.getCallerIdentity();
      account = identity.account();
    }

    String body = "{\"status\":\"ok\",\"account\":\"" + account + "\"}";

    return new APIGatewayProxyResponseEvent().withStatusCode(200).withBody(body);
  }
}
