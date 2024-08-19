import {
  APIGatewayProxyEvent,
  APIGatewayProxyResult,
  Context,
} from 'aws-lambda';

import AWS from 'aws-sdk';

const s3 = new AWS.S3();

// eslint-disable-next-line @typescript-eslint/no-unused-vars
exports.handler = async (_event: APIGatewayProxyEvent, _context: Context) => {
  console.info('Serving lambda request.');

  const result = await s3.listBuckets().promise();

  const response: APIGatewayProxyResult = {
    statusCode: 200,
    body: `Hello lambda - found ${result.Buckets?.length || 0} buckets`,
  };
  return response;
};
