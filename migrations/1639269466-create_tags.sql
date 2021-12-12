create table tags (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id),

    name text not null,

    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
create index on tags (user_id, name);
select manage_updated_at('tags');
