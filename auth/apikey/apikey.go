package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"

	"github.com/gofrs/uuid"
	"github.com/metagram-net/firehose/db"
)

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

func (p Plaintext) Hash() []byte {
	b := sha256.Sum256(p.bytes)
	return b[:]
}

func generate() (Plaintext, error) {
	b := make([]byte, keyLength)
	_, err := rand.Read(b)
	if err != nil {
		return Plaintext{}, err
	}
	return Plaintext{bytes: b}, nil
}

func Create(ctx context.Context, q db.Querier, name string, userID uuid.UUID) (*Plaintext, *db.ApiKey, error) {
	plain, err := generate()
	if err != nil {
		return nil, nil, err
	}

	k, err := q.ApiKeyCreate(ctx, db.ApiKeyCreateParams{
		Name:         name,
		UserID:       userID,
		HashedSecret: plain.Hash(),
	})
	if err != nil {
		return nil, nil, err
	}
	return &plain, &k, nil
}

func Find(ctx context.Context, q db.Querier, userID uuid.UUID, plain Plaintext) (*db.ApiKey, error) {
	k, err := q.ApiKeyFind(ctx, db.ApiKeyFindParams{
		UserID:       userID,
		HashedSecret: plain.Hash(),
	})
	if err != nil {
		return nil, err
	}
	return &k, nil
}
