create function require_migration(mid integer) returns void as $$
declare
    mrow schema_migrations%rowtype;
begin
    select * into mrow from schema_migrations where id = mid;
    if not found then
        raise exception 'Required migration has not been run: %', mid;
    end if;
end;
$$ language plpgsql;
