package moray

import (
	"fmt"
	"strings"

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
	return "UUID?"
}

type String string

func (s *String) String() string {
	if s == nil {
		return "<nil>"
	}
	return string(*s)
}

func (s *String) Set(str string) error {
	*s = String(str)
	return nil
}

func (*String) Type() string {
	return "string?"
}

type UUIDs []uuid.UUID

func (us *UUIDs) String() string {
	if us == nil {
		return "<nil>"
	}

	var s []string
	for _, u := range *us {
		s = append(s, u.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ", "))
}

func (us *UUIDs) Type() string {
	return "[UUID]"
}

func (us *UUIDs) Set(s string) error {
	u, err := uuid.FromString(s)
	*us = append(*us, u)
	return err
}
