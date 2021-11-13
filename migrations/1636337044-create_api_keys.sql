create table api_keys (
    id uuid primary key default gen_random_uuid(),
    user_id uuid references users(id),

    name text not null,
    hashed_secret bytea unique not null,

    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
select manage_updated_at('api_keys');
