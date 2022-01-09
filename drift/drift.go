package drift

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"text/tabwriter"
	"text/template"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/jackc/pgconn"
	"github.com/metagram-net/firehose/db"
)

var (
	ErrNegativeID  = errors.New("migration ID must not be negative")
	ErrDuplicateID = errors.New("duplicate migration ID")
)

// A MigrationID is a nonnegative integer that will be used to sort migrations.
//
// This will often be a Unix timestamp in seconds, so it's represented as as an
// int64 for easy conversion. That technically allows negative numbers
// (although getting one in modern times would be concerning!), so use
// NewMigrationID to check for negative values.
type MigrationID int64

func NewMigrationID(i int64) (MigrationID, error) {
	if i < 0 {
		return 0, fmt.Errorf("%w: %d", ErrNegativeID, i)
	}
	return MigrationID(i), nil
}

func (*MigrationID) Type() string {
	return "non_negative_integer"
}

func (m *MigrationID) String() string {
	if m == nil {
		return ""
	}
	return strconv.Itoa(int(*m))
}

func (m *MigrationID) Set(s string) error {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("not a valid integer: %w", err)
	}
	id, err := NewMigrationID(i)
	*m = id
	return err
}

func (m MigrationID) Width() int {
	return len(fmt.Sprintf("%d", m))
}

func mustID(s string) MigrationID {
	var id MigrationID
	if err := id.Set(s); err != nil {
		panic(err)
	}
	return id
}

// Migrate runs all unapplied migrations in ID order, least to greatest. It
// skips any migrations that have already been applied.
func Migrate(ctx context.Context, db *sql.DB, migrationsDir string) error {
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
		log.Printf("Applying %s", f.Name)
		if err := apply(ctx, db, f); err != nil {
			return err
		}
	}
	log.Print("All migrations applied")
	return nil
}

type migrationRecord struct {
	ID    MigrationID `db:"id"`
	Slug  string      `db:"slug"`
	RunAt time.Time   `db:"run_at"`
}

var qApplied, _ = sq.Select("*").From("schema_migrations").OrderBy("id asc").MustSql()

func applied(db *sql.DB) ([]migrationRecord, error) {
	rows, err := db.Query(qApplied)
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "42P01" { // undefined_table
		// The expected table doesn't exist. This is almost certainly because
		// we haven't run the first migration that will create this table.
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ms []migrationRecord
	return ms, scan.RowsStrict(&ms, rows)
}

// reFilename matches the migration filename convention.
//
// Some examples of names:
//
//  - 0-init.sql
//  - 1234567890-create_users.sql"
//
var reFilename = regexp.MustCompile(`^(?P<id>\d+)-(?P<slug>.*)\.sql$`)

type migrationFile struct {
	Path    string
	Name    string
	Content string

	ID   MigrationID
	Slug string

	idRaw string
}

// TODO: Use an afero.Fs to make this easier to test.

func available(dir string) ([]migrationFile, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not list migration files: %w", err)
	}

	var ms []migrationFile
	for _, f := range files {
		name := f.Name()
		m := reFilename.FindStringSubmatch(name)
		if m == nil {
			// TODO: Only log this in verbose mode
			log.Printf("Skipping non-migration file: %s", name)
			continue
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

			// The subexpression cannot match negative integers, so this can
			// only fail if the ID doesn't fit into an int64.
			ID:   mustID(m[reFilename.SubexpIndex("id")]),
			Slug: m[reFilename.SubexpIndex("slug")],

			idRaw: m[reFilename.SubexpIndex("id")],
		})
	}

	seen := make(map[MigrationID]migrationFile)
	for _, m := range ms {
		if other, ok := seen[m.ID]; ok {
			return nil, fmt.Errorf("%w: %s, %s", ErrDuplicateID, other.Name, m.Name)
		}
		seen[m.ID] = m
	}
	return ms, nil
}

func diff(applied []migrationRecord, files []migrationFile) []migrationFile {
	skip := make(map[MigrationID]struct{})
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

func apply(ctx context.Context, db *sql.DB, f migrationFile) error {
	if skipTx(f.Content) {
		return run(ctx, db, f.Content)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := claim(ctx, tx, f.ID, f.Slug); err != nil {
		return err
	}
	if err := run(ctx, tx, f.Content); err != nil {
		return err
	}
	return tx.Commit()
}

// reNoTxComment finds the `--drift::no-transaction` directive as a one-line
// SQL comment.
var reNoTxComment = regexp.MustCompile(`(?m)^--drift:no-transaction`)

func skipTx(content string) bool {
	return reNoTxComment.MatchString(content)
}

func claim(ctx context.Context, tx db.DBTX, id MigrationID, slug string) error {
	query, args, err := db.Pq.Select().
		Column("_drift_claim_migration("+sq.Placeholders(2)+")", id, slug).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func run(ctx context.Context, tx db.DBTX, content string) error {
	_, err := tx.ExecContext(ctx, content)
	return err
}

// Setup creates the "init" migration that will prepare the database for
// migrations. This will create the migrations directory if needed.
func Setup(migrationsDir string) (string, error) {
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

// NewFile creates a new migration file with a placeholder comment in it.
func NewFile(migrationsDir string, id MigrationID, slug string, tmpl *template.Template) (string, error) {
	if tmpl == nil {
		tmpl = defaultTemplate
	}

	if id == -1 {
		var err error
		ts := time.Now().Unix()
		id, err = NewMigrationID(ts)
		if err != nil {
			return "", fmt.Errorf("invalid migration ID: %w", err)
		}
	}

	files, err := available(migrationsDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if f.ID == id {
			return "", fmt.Errorf("%w: %d: %s", ErrDuplicateID, id, f.Name)
		}
	}

	slug = slugify(slug)
	name := filename(idWidth(files), id, slug)
	path := filepath.Join(migrationsDir, name)
	data := TemplateData{
		ID:   id,
		Slug: slug,
	}

	//#nosec G306 // Normal permissions for non-sensitive files.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	return path, tmpl.Execute(f, data)
}

//go:embed new.sql
var newContent string
var defaultTemplate = template.Must(template.New("default").Parse(newContent))

type TemplateData struct {
	ID   MigrationID
	Slug string
}

// reSeparator matches runs of common characters types as separators in
// interactive command-line usage.
var reSeparator = regexp.MustCompile(`[\-\s._/]+`)

func slugify(s string) string {
	return reSeparator.ReplaceAllString(s, "_")
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

//go:embed init.sql
var initContent string

func Renumber(dir string) error {
	files, err := available(dir)
	if err != nil {
		return err
	}
	width := idWidth(files)

	type rename struct {
		from string
		to   string
	}
	var renames []rename
	for _, f := range files {
		id := f.idRaw
		if len(id) < width {
			renames = append(renames, rename{
				from: f.Name,
				to:   filename(width, f.ID, f.Slug),
			})
		}
	}

	table := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	log.Printf("Renames:")
	for _, r := range renames {
		fmt.Fprintf(table, "%s\t->\t%s\n", r.from, r.to)
	}
	table.Flush()

	for _, r := range renames {
		old := filepath.Join(dir, r.from)
		new := filepath.Join(dir, r.to)
		if err := os.Rename(old, new); err != nil {
			return err
		}
	}
	return nil
}

func idWidth(files []migrationFile) int {
	width := 0
	for _, f := range files {
		w := f.ID.Width()
		if w > width {
			width = w
		}
	}
	return width
}

func filename(idWidth int, id MigrationID, slug string) string {
	return fmt.Sprintf("%0*d-%s.sql", idWidth, id, slug)
}
