package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run migrations",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			if err := migrate(viper.GetString("migrations-dir")); err != nil {
				log.Fatal(err.Error())
			}
		},
	}
	return cmd
}

func migrate(migrationsDir string) error {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer db.Close()

	// 1. select * from schema_migrations
	records, err := applied(db)
	if err != nil {
		return fmt.Errorf("could not get applied migrations: %w", err)
	}

	// 2. ls migrations_dir
	files, err := available(migrationsDir)
	if err != nil {
		return fmt.Errorf("could not get available migrations: %w", err)
	}

	// 3. diff IDs
	needed := diff(records, files)
	for _, f := range needed {
		log.Printf("applying %s", f.Name)
		if err := apply(db, f); err != nil {
			return err
		}
	}

	return nil
}

type migrationRecord struct {
	ID    int       `db:"id"`
	Slug  string    `db:"slug"`
	RunAt time.Time `db:"run_at"`
}

type migrationFile struct {
	Path    string
	Name    string
	Content string

	ID   int
	Slug string
}

var qApplied, _ = sq.Select("*").From("schema_migrations").OrderBy("id asc").MustSql()

func applied(db *sql.DB) ([]migrationRecord, error) {
	rows, err := db.Query(qApplied)
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "42P01" { // undefined_table
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ms []migrationRecord
	return ms, scan.RowsStrict(&ms, rows)
}

var reFname = regexp.MustCompile(`^(?P<id>\d+)-(?P<slug>.*)\.sql$`)

func mustInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

var ErrDuplicateID = errors.New("duplicate migration ID")

func available(dir string) ([]migrationFile, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not list migration files: %w", err)
	}

	var ms []migrationFile
	for _, f := range files {
		name := f.Name()
		m := reFname.FindStringSubmatch(name)
		if m == nil {
			return nil, fmt.Errorf("filename does not fit migration pattern: %s", name)
		}
		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		ms = append(ms, migrationFile{
			Path:    path,
			Name:    name,
			Content: string(content),

			ID:   mustInt(m[reFname.SubexpIndex("id")]),
			Slug: m[reFname.SubexpIndex("slug")],
		})
	}

	seen := make(map[int]migrationFile)
	for _, m := range ms {
		if other, ok := seen[m.ID]; ok {
			return nil, fmt.Errorf("%w: %s, %s", ErrDuplicateID, other.Name, m.Name)
		}
		seen[m.ID] = m
	}
	return ms, nil
}

func diff(applied []migrationRecord, files []migrationFile) []migrationFile {
	skip := make(map[int]struct{})
	for _, r := range applied {
		skip[r.ID] = struct{}{}
	}

	var needed []migrationFile
	for _, f := range files {
		if _, ok := skip[f.ID]; ok {
			continue
		}
		needed = append(needed, f)
	}

	sort.Slice(needed, func(i, j int) bool { return needed[i].ID < needed[j].ID })
	return needed
}

func apply(db *sql.DB, f migrationFile) error {
	if skipTx(f.Content) {
		return run(db, f.Content)
	}

	// TODO: thread a context through here so this can be canceled cleanly
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := claim(tx, f.ID, f.Slug); err != nil {
		return err
	}
	if err := run(tx, f.Content); err != nil {
		return err
	}
	return tx.Commit()
}

func applyNoTx(db *sql.DB, f migrationFile) error {
	return run(db, f.Path)
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type execable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func claim(tx execable, id int, slug string) error {
	query, args, err := psql.Select().
		Column("_drift_claim_migration("+sq.Placeholders(2)+")", id, slug).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query, args...)
	return err
}

func run(tx execable, content string) error {
	_, err := tx.Exec(content)
	return err
}

// reNoTxComment finds the "drift::no-transaction" directive at the beginning
// of a one-line SQL comment.
var reNoTxComment = regexp.MustCompile(`(?m)^--drift:no-transaction`)

func skipTx(content string) bool {
	return reNoTxComment.MatchString(content)
}
