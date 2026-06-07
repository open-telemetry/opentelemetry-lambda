import { STSClient, GetCallerIdentityCommand } from '@aws-sdk/client-sts';

const sts = new STSClient({});

export const handler = async () => {
  const identity = await sts.send(new GetCallerIdentityCommand({}));
  return {
    statusCode: 200,
    body: JSON.stringify({ status: 'ok', account: identity.Account }),
  };
};
