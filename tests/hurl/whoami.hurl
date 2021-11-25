GET {{root_url}}/auth/whoami

HTTP/* 401

[Asserts]
jsonpath "$.error_code" == "missing_authorization"
jsonpath "$.error_message" == "The Authorization header was missing or the wrong format."


GET {{root_url}}/auth/whoami
Authorization: Bearer a

HTTP/* 401

[Asserts]
jsonpath "$.error_code" == "invalid_authorization"
jsonpath "$.error_message" == "The provided credentials were not valid."


GET {{root_url}}/auth/whoami
Authorization: Basic a

HTTP/* 401

[Asserts]
jsonpath "$.error_code" == "invalid_authorization"
jsonpath "$.error_message" == "The provided credentials were not valid."


GET {{root_url}}/auth/whoami
Authorization: Basic {{basic_auth}}

HTTP/* 200

[Asserts]
jsonpath "$.id" exists
