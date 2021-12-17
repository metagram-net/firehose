package types

// TODO: I really really don't like `types.*`. Move this somewhere else.

//go:generate enumer -output=drop_status_enumer.go -type=DropStatus -json -sql -linecomment

type DropStatus int

const (
	StatusUnknown DropStatus = iota // unknown
	StatusUnread                    // unread
	StatusRead                      // read
	StatusSaved                     // saved
)
