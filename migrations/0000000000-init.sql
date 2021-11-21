--drift:no-transaction

begin;

create table schema_migrations (
    id integer primary key,
    slug text not null,
    run_at timestamp not null default current_timestamp
);

create function _drift_claim_migration(mid integer, mslug text) returns void as $$
    insert into schema_migrations (id, slug) values (mid, mslug);
$$ language sql;

-- Normally, this would be the first thing in the migration, but we had to
-- create the schema_migrations table first!
select _drift_claim_migration(0, 'init');

commit;
