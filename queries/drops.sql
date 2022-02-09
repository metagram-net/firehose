-- name: DropFind :one
select * from drops
where user_id = $1 and id = $2;

-- name: DropNext :one
select * from drops
where user_id = $1 and status = 'unread'
order by moved_at asc;

-- name: DropList :many
select * from drops
where user_id = $1 and status = ANY(@statuses::drop_status[])
order by moved_at asc
limit $2;

-- name: DropCreate :one
insert into drops
(user_id, title, url, status, moved_at)
values ($1, $2, $3, $4, $5)
returning *;

-- name: DropMove :one
update drops
set status = $3, moved_at = $4
where user_id = $1 and id = $2
returning *;

-- name: DropDelete :one
delete from drops where user_id = $1 and id = $2 returning *;

-- custom: DropUpdate
