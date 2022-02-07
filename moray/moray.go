package moray

import (
	"github.com/gofrs/uuid"
)

type UUID uuid.UUID

func (u *UUID) String() string {
	if u == nil {
		return "<nil>"
	}
	return (*uuid.UUID)(u).String()
}

func (u *UUID) Set(s string) error {
	uu, err := uuid.FromString(s)
	*u = UUID(uu)
	return err
}

func (u *UUID) Type() string {
	return "UUID"
}
