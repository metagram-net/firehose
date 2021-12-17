-- name: UserFind :one
select * from users where id = $1;

-- name: UserCreate :one
insert into users (email_address) values ($1) returning *;
