create table users (
    id uuid primary key default gen_random_uuid(),
    email_address text unique not null,

    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
);
select manage_updated_at('users');
