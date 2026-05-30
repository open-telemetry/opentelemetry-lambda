require "aws-sdk-core"
require "json"

def handler(event:, context:)
  identity = Aws::STS::Client.new.get_caller_identity
  {
    statusCode: 200,
    body: JSON.generate({ status: "ok", account: identity.account }),
  }
end
