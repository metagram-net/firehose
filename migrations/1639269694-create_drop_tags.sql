create table drop_tags (
    id uuid primary key default gen_random_uuid(),

    drop_id uuid not null references drops(id),
    tag_id uuid not null references tags(id),

    created_at timestamp not null default current_timestamp
);
create index on drop_tags (drop_id);
create index on drop_tags (tag_id);
