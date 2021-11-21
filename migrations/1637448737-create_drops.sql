create table drops (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id),

    title text,
    url text not null,
    status text not null,
    moved_at timestamp not null,

    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
create index on drops (user_id, status);
select manage_updated_at('drops');
