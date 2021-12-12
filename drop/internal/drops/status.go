package drops

//go:generate enumer -type=Status -json -sql -linecomment

type Status int

const (
	StatusUnknown Status = iota // unknown
	StatusUnread                // unread
	StatusRead                  // read
	StatusSaved                 // saved
)
