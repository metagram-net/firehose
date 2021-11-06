package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setup",
		Aliases: []string{"init"},
		Short:   "Set up the migrations directory",
		Args:    cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			path, err := setup(viper.GetString("migrations-dir"))
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Printf("Wrote %s\n", path)
		},
	}
	return cmd
}

func setup(migrationsDir string) (string, error) {
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		return "", fmt.Errorf("could not create migrations directory: %w", err)
	}
	name := fmt.Sprintf("%d-%s.sql", 0, "init")
	path := filepath.Join(migrationsDir, name)
	if err := safeWriteFile(path, []byte(initContent), 0o644); err != nil {
		return "", fmt.Errorf("could not create migration file: %w", err)
	}
	return path, nil
}

// safeWriteFile is like os.WriteFile but it fails if the file already exists.
func safeWriteFile(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	// Prefer the write error over the close error.
	_, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}

// TODO: Put this in a template file and embed it.
const initContent = `/*
Set up the Drift framework requirements. Naturally, this first migration is
going to break a few rules ;)

First, this includes a drift:no-transaction directive, which tells Drift to
skip two steps it would normally take:
1. Opening a transaction around the migration file. In Postgres, DDL can be
   done in a transaction. This can make some migrations safer, so Drift assumes
   transactions as the default.
2. Calling _drift_claim_migration(id, slug) before running the file. Since this
   claim would fail on a duplicate ID, this ensures we never run a migration
   twice (since it's normally part of a transaction).

It doesn't make sense to call _drift_claim_migration yet, because this is the
migration that defines it!

You can modify the _drift_claim_migration function if you want to. The only
expectation Drift has of it (besides the signature) is that it writes the
migration ID to the table and fails if that ID is already recorded.

You can also modify the schema_migrations table, but (at least for now) Drift
assumes that the migration records table has exactly that name and has the
integer primary key id column.
*/
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
`
