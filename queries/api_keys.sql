-- name: ApiKeyFind :one
select * from api_keys where user_id = $1 and hashed_secret = $2;

-- name: ApiKeyCreate :one
insert into api_keys (name, user_id, hashed_secret) values ($1, $2, $3) returning *;
