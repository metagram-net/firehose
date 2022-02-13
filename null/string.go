package null

import (
	"bytes"
	"database/sql"
	"encoding/json"
)

type String struct {
	Value   string
	Present bool
}

func (s *String) Ptr() *string {
	if s.Present {
		// Copy the string so the reference is distinct from the contained
		// value.
		v := s.Value
		return &v
	}
	return nil
}

func (s *String) SQL() sql.NullString {
	return sql.NullString{
		String: s.Value,
		Valid:  s.Present,
	}
}

// Implement pflag.Value

func (s *String) String() string {
	if s.Present {
		return s.Value
	}
	return "<nil>"
}

//nolint:unparam // pflag.Value needs this to return an error.
func (s *String) Set(str string) error {
	*s = String{str, true}
	return nil
}

func (*String) Type() string {
	return "Optional<String>"
}

// Implement json.Marshaler and json.Unmarshaler

func (s String) MarshalJSON() ([]byte, error) {
	if s.Present {
		return json.Marshal(s.Value)
	}
	return json.Marshal(nil)
}

func (s *String) UnmarshalJSON(b []byte) error {
	// By convention, this is a no-op.
	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	s.Present = true
	return json.Unmarshal(b, &s.Value)
}
