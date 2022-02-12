-- name: DropTagApply :one
insert into drop_tags
(drop_id, tag_id)
values ($1, $2)
returning *;

-- name: DropTagsIntersect :many
delete from drop_tags
where drop_id = $1 and tag_id != any(@tag_ids::uuid[])
returning *;

-- custom: DropTagsApply
