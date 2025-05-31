import {
  APIGatewayProxyEvent,
  APIGatewayProxyResult,
  Context,
} from 'aws-lambda';

import { STSClient, GetCallerIdentityCommand } from '@aws-sdk/client-sts';

const sts = new STSClient({});

export const handler = async (
  event: APIGatewayProxyEvent,
  _context: Context,
): Promise<APIGatewayProxyResult> => {
  console.log('Received event:', JSON.stringify(event, null, 2));
  console.log('Received context:', JSON.stringify(_context, null, 2));

  try {
    const result = await sts.send(new GetCallerIdentityCommand({}));

    const response: APIGatewayProxyResult = {
      statusCode: 200,
      body: JSON.stringify({
        message: 'Caller identity retrieved successfully',
        identity: {
          Account: result.Account,
          Arn: result.Arn,
          UserId: result.UserId,
        },
      }),
    };
    return response;
  } catch (error) {
    console.error('Error retrieving caller identity:', error);
    return {
      statusCode: 500,
      body: 'Internal Server Error',
    };
  }
};
