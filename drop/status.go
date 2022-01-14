package drop

import (
	"fmt"

	"github.com/metagram-net/firehose/db"
)

//go:generate enumer -type=Status -json -linecomment

type Status int

const (
	StatusUnread Status = iota + 1 // unread
	StatusRead                     // read
	StatusSaved                    // saved
)

// StatusValueStrings returns all valid values of the enum as strings.
func StatusValueStrings() []string {
	return []string{
		StatusUnread.String(),
		StatusRead.String(),
		StatusSaved.String(),
	}
}

// TODO: Use a linter to make sure these are exhaustive switches.

func StatusModel(s db.DropStatus) Status {
	switch s {
	case db.DropStatusUnread:
		return StatusUnread
	case db.DropStatusRead:
		return StatusRead
	case db.DropStatusSaved:
		return StatusSaved
	default:
		panic(fmt.Sprintf("unknown status: %s", s))
	}
}

func (s Status) Model() db.DropStatus {
	switch s {
	case StatusUnread:
		return db.DropStatusUnread
	case StatusRead:
		return db.DropStatusRead
	case StatusSaved:
		return db.DropStatusSaved
	default:
		panic(fmt.Sprintf("unknown status: %s", s))
	}
}
