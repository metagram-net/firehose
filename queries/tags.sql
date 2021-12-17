-- name: TagFind :one
select * from tags where user_id = $1 and id = $2;

-- name: TagFindAll :many
select * from tags where user_id = $1 and id = ANY(@ids::uuid[]);

-- name: TagList :many
select * from tags where user_id = $1;

-- name: TagsDrop :many
select tags.* from tags
join drop_tags on drop_tags.tag_id = tags.id
join drops on drops.id = drop_tags.drop_id
where drops.user_id = $1 and drops.id = $2;

-- name: TagsDrops :many
select tags.*, drops.id as drop_id from tags
join drop_tags on drop_tags.tag_id = tags.id
join drops on drops.id = drop_tags.drop_id
where drops.user_id = $1 and drops.id = ANY(@drop_ids::uuid[]);

-- name: TagCreate :one
insert into tags (user_id, name) values ($1, $2) returning *;

-- name: TagMove :one
update tags set name = $3 where user_id = $1 and id = $2 returning *;

-- name: TagDelete :one
delete from tags where user_id = $1 and id = $2 returning *;
