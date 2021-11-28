GET {{root_url}}/does-not-exist

HTTP/* 404

[Asserts]
jsonpath "$.error_code" == "not_found"
jsonpath "$.error_message" == "The requested route or resource does not exist."


POST {{root_url}}/.well-known/health-check

HTTP/* 405

[Asserts]
jsonpath "$.error_code" == "method_not_allowed"
jsonpath "$.error_message" == "The requested HTTP method cannot be handled by this route."
