package main

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrNegative = errors.New("must not be negative")

type migrationID int64

func newMigrationID(i int64) (migrationID, error) {
	if i < 0 {
		return 0, fmt.Errorf("%w: %d", ErrNegative, i)
	}
	return migrationID(i), nil
}

func (m *migrationID) String() string {
	if m == nil {
		return ""
	}
	return strconv.Itoa(int(*m))
}

func (m *migrationID) Set(s string) error {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("not a valid integer: %w", err)
	}
	id, err := newMigrationID(i)
	*m = id
	return err
}

func (*migrationID) Type() string {
	return "non_negative_integer"
}
