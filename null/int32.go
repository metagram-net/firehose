package null

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type Int32 struct {
	Value   int32
	Present bool
}

func (i *Int32) Ptr() *int32 {
	if i.Present {
		v := i.Value
		return &v
	}
	return nil
}

// Implement pflag.Value

func (i *Int32) String() string {
	if i.Present {
		return strconv.Itoa(int(i.Value))
	}
	return "<nil>"
}

func (i *Int32) Set(s string) error {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	*i = Int32{int32(n), true}
	return nil
}

func (*Int32) Type() string {
	return "Optional<String>"
}

// Implement json.Marshaler and json.Unmarshaler

func (i Int32) MarshalJSON() ([]byte, error) {
	if i.Present {
		return json.Marshal(i.Value)
	}
	return json.Marshal(nil)
}

func (i *Int32) UnmarshalJSON(b []byte) error {
	// By convention, this is a no-op.
	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	i.Present = true
	return json.Unmarshal(b, &i.Value)
}
