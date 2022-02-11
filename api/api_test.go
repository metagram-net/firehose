package api_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/metagram-net/firehose/api"
)

type Custom struct {
	UUID uuid.UUID
}

func (c *Custom) FromParam(s string) error {
	u, err := uuid.FromString(s)
	if err != nil {
		return err
	}
	(*c).UUID = u
	return nil
}

func TestFromVars(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type Params struct {
			Ignored string `json:"ignored"`

			Bool bool `var:"bool"`

			String string `var:"string"`
			Bytes  []byte `var:"bytes"`

			Int   int   `var:"int"`
			Int8  int8  `var:"int8"`
			Int16 int16 `var:"int16"`
			Int32 int32 `var:"int32"`
			Int64 int64 `var:"int64"`

			Uint   uint   `var:"uint"`
			Uint8  uint8  `var:"uint8"`
			Uint16 uint16 `var:"uint16"`
			Uint32 uint32 `var:"uint32"`
			Uint64 uint64 `var:"uint64"`

			Uintptr uintptr `var:"uintptr"`

			Byte byte `var:"byte"`
			Rune rune `var:"rune"`

			Float32 float32 `var:"float32"`
			Float64 float64 `var:"float64"`

			Complex64  complex64  `var:"complex64"`
			Complex128 complex128 `var:"complex128"`

			Custom Custom `var:"custom"`
		}

		vars := map[string]string{
			"ignored": "do not use",

			"bool": "true",

			"string": "strstr",
			"bytes":  "⛄",

			"int":   "-1",
			"int8":  "8",
			"int16": "16",
			"int32": "32",
			"int64": "64",

			"uint":   "2",
			"uint8":  "7",
			"uint16": "15",
			"uint32": "31",
			"uint64": "63",

			"uintptr": "1000",

			"byte": "9",
			"rune": "97",

			"float32": "3.14",
			"float64": "-6.28",

			"complex64":  "-1+2i",
			"complex128": "3-4i",

			"custom": "fd762a80-e331-44da-8b26-c6b014a4d953",
		}

		expected := Params{
			Ignored: "",

			Bool: true,

			String: "strstr",
			Bytes:  []byte{0xe2, 0x9b, 0x84},

			Int:   -1,
			Int8:  8,
			Int16: 16,
			Int32: 32,
			Int64: 64,

			Uint:   2,
			Uint8:  7,
			Uint16: 15,
			Uint32: 31,
			Uint64: 63,

			Uintptr: 1000,

			Byte: 9,
			Rune: 'a',

			Float32: 3.14,
			Float64: -6.28,

			Complex64:  -1 + 2i,
			Complex128: 3 - 4i,

			Custom: Custom{uuid.FromStringOrNil("fd762a80-e331-44da-8b26-c6b014a4d953")},
		}

		var p Params
		err := api.FromVars(vars, &p)
		require.NoError(t, err)
		assert.Equal(t, expected, p)
	})

	t.Run("errors", func(t *testing.T) {
		t.Skip("TODO")
	})
}
