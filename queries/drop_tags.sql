-- name: DropTagApply :one
insert into drop_tags
(drop_id, tag_id)
values ($1, $2)
returning *;
