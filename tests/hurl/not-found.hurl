GET {{root_url}}/does-not-exist

HTTP/* 404

[Asserts]
jsonpath "$.error_code" == "not_found"
jsonpath "$.error_message" == "The requested route or resource does not exist."
