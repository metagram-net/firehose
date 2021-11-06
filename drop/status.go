package drop

//go:generate enumer -type=Status -json -sql -linecomment

type Status int

const (
	StatusUnread Status = iota // unread
	StatusRead                 // read
	StatusSaved                // saved
)
