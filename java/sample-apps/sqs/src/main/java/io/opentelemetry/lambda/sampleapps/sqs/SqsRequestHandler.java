package io.opentelemetry.lambda.sampleapps.sqs;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.SQSEvent;
import com.amazonaws.services.lambda.runtime.events.SQSEvent.SQSMessage;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

public class SqsRequestHandler implements RequestHandler<SQSEvent, Void> {

  private static final Logger logger = LogManager.getLogger(SqsRequestHandler.class);

  @Override
  public Void handleRequest(SQSEvent event, Context context) {
    logger.info("Processing message(s) from SQS");

    for (SQSMessage msg : event.getRecords()) {
      logger.info("SOURCE QUEUE: " + msg.getEventSourceArn());
      logger.info("MESSAGE: " + msg.getBody());
    }

    return null;
  }
}
