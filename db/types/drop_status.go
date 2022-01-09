package types

// TODO: I really really don't like `types.*`. Move this somewhere else.

//go:generate enumer -output=drop_status_enumer.go -type=DropStatus -json -sql -linecomment

type DropStatus int

const (
	StatusUnread DropStatus = iota + 1 // unread
	StatusRead                         // read
	StatusSaved                        // saved
)

// DropStatusValueStrings returns all valid values of the enum as strings.
func DropStatusValueStrings() []string {
	return []string{
		StatusUnread.String(),
		StatusRead.String(),
		StatusSaved.String(),
	}
}
