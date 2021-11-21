package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/blockloop/scan"
	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/db"
)

type Record struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	Name         string    `db:"name"`
	HashedSecret []byte    `db:"hashed_secret"`
	db.Timestamps
}

const keyLength = 32 // bytes

// Douglas Crockford's Base32 variant: https://www.crockford.com/base32.html
var crock32 = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(base32.NoPadding)

type Plaintext struct {
	bytes []byte
}

func NewPlaintext(s string) (Plaintext, error) {
	b, err := crock32.DecodeString(s)
	if err != nil {
		return Plaintext{}, err
	}
	return Plaintext{bytes: b}, nil
}

func (p Plaintext) String() string {
	return crock32.EncodeToString(p.bytes)
}

func (p Plaintext) Hash() [32]byte {
	return sha256.Sum256(p.bytes)
}

func (p Plaintext) byteaHex() string {
	hash := p.Hash()
	return fmt.Sprintf("\\x%x", hash)
}

func generate() (Plaintext, error) {
	b := make([]byte, keyLength)
	_, err := rand.Read(b)
	if err != nil {
		return Plaintext{}, err
	}
	return Plaintext{bytes: b}, nil
}

func Create(ctx context.Context, tx db.Queryable, name string, userID uuid.UUID) (*Plaintext, *Record, error) {
	plain, err := generate()
	if err != nil {
		return nil, nil, err
	}

	query, args, err := db.Pq.
		Insert("api_keys").
		SetMap(map[string]interface{}{
			"user_id":       userID,
			"name":          name,
			"hashed_secret": plain.byteaHex(),
		}).
		Suffix("returning *").
		ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}

	var r Record
	if err := scan.RowStrict(&r, rows); err != nil {
		return nil, nil, err
	}
	return &plain, &r, nil
}

func Find(ctx context.Context, tx db.Queryable, userID uuid.UUID, plain Plaintext) (*Record, error) {
	query, args, err := db.Pq.
		Select("*").
		From("api_keys").
		Where(sq.Eq{
			"user_id":       userID,
			"hashed_secret": plain.byteaHex(),
		}).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var r Record
	return &r, scan.RowStrict(&r, rows)
}

func Delete(ctx context.Context, tx db.Queryable, id uuid.UUID) error {
	query, args, err := db.Pq.
		Delete("api_keys").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func DeleteByPlaintext(ctx context.Context, tx db.Queryable, plain Plaintext) error {
	query, args, err := db.Pq.
		Delete("api_keys").
		Where(sq.Eq{"hashed_secret": plain.byteaHex()}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, args...)
	return err
}
