package drop

//go:generate enumer -type=Status -json -sql -linecomment

type Status int

const (
	// TODO: Start enum at 1
	StatusUnread Status = iota // unread
	StatusRead                 // read
	StatusSaved                // saved
)
