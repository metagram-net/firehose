package apitest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" // database/sql driver: pgx
	"github.com/metagram-net/firehose/api"
	"github.com/metagram-net/firehose/auth/user"
	"github.com/metagram-net/firehose/clock"
	"github.com/metagram-net/firehose/db"
)

func Context(t *testing.T, timeout time.Duration) context.Context {
	if timeout <= 0 {
		return context.Background()
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

func Clock(_ *testing.T) clock.Frozen {
	// This is the timestamp for the first commit in this repo.
	ref := time.Date(2021, 11, 1, 20, 7, 16, 0, time.FixedZone("PDT", -7*60*60))
	return clock.Freeze(ref.UTC())
}

func Tx(t *testing.T, ctx context.Context) *sql.Tx {
	db, err := sql.Open("pgx", mustEnv(t, "TEST_DATABASE_URL"))
	if err != nil {
		t.Fatalf("Could not establish database connection: %s", err.Error())
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Could not start transaction: %s", err.Error())
	}
	t.Cleanup(func() {
		if envBool(t, "TEST_DATABASE_COMMIT") {
			if err := tx.Commit(); err != nil {
				t.Logf("Could not commit transaction: %s", err.Error())
			}
		} else {
			if err := tx.Rollback(); err != nil {
				t.Logf("Could not roll back transaction: %s", err.Error())
			}
		}
	})
	return tx
}

func User(t *testing.T, ctx context.Context, tx db.Queryable) api.User {
	email := fmt.Sprintf("%x@user.test", UUID(t))
	r, err := user.Create(ctx, tx, email)
	if err != nil {
		t.Fatalf("Could not create user: %s", err.Error())
	}
	return api.User{ID: r.ID}
}

func UUID(t *testing.T) uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("Could not generate UUID: %s", err.Error())
	}
	return id
}

func mustEnv(t *testing.T, name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		skipf(t, "Environment variable not set: %s", name)
	}
	if val == "" {
		skipf(t, "Environment variable was blank: %s", name)
	}
	return val
}

func skipf(t *testing.T, format string, args ...interface{}) {
	if isCI(t) {
		t.Fatalf(format, args...)
	} else {
		t.Skipf(format, args...)
	}
}

func isCI(t *testing.T) bool { return envBool(t, "CI") }

func envBool(t *testing.T, name string) bool {
	val, ok := os.LookupEnv(name)
	if !ok {
		t.Logf("%s not set, assuming false", name)
		return false
	}
	bval, err := strconv.ParseBool(val)
	if err != nil {
		t.Logf("Could not parse %s value, assuming false: %s", name, val)
		return false
	}
	return bval
}
