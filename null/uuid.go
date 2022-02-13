package null

import (
	"bytes"
	"encoding/json"

	"github.com/gofrs/uuid"
)

type UUID struct {
	Value   uuid.UUID
	Present bool
}

// Implement pflag.Value

func (u *UUID) String() string {
	if u.Present {
		return u.Value.String()
	}
	return ""
}

func (u *UUID) Set(s string) error {
	uu, err := uuid.FromString(s)
	if err != nil {
		return err
	}
	*u = UUID{uu, true}
	return nil
}

func (*UUID) Type() string {
	return "UUID"
}

// Implement json.Marshaler and json.Unmarshaler

func (u UUID) MarshalJSON() ([]byte, error) {
	if u.Present {
		return json.Marshal(u.Value)
	}
	return json.Marshal(nil)
}

func (u *UUID) UnmarshalJSON(b []byte) error {
	// By convention, this is a no-op.
	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	u.Present = true
	return json.Unmarshal(b, &u.Value)
}
