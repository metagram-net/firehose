package moray

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func BindFlags(cmd *cobra.Command, args interface{}) {
	flags := cmd.Flags()

	vv := reflect.ValueOf(args).Elem()
	t := vv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		val := f.Tag.Get("flag")
		if val == "" {
			continue
		}
		tag := parseTag(val)

		usage := f.Tag.Get("usage")

		dest := vv.FieldByIndex(f.Index).Addr().Interface()
		switch dest := dest.(type) {
		case pflag.Value:
			flags.Var(dest, tag.Name, usage)
		case *int32:
			flags.Int32Var(dest, tag.Name, 0, usage)
		default:
			panic(fmt.Sprintf("unhandled destination type: %T", dest))
		}

		if tag.Required {
			cmd.MarkFlagRequired(tag.Name)
		}
	}
}

type Tag struct {
	Name     string
	Required bool
}

func parseTag(tag string) Tag {
	parts := strings.Split(tag, ",")
	t := Tag{Name: parts[0]}
	for _, opt := range parts[1:] {
		switch opt {
		case "required":
			t.Required = true
		default:
			panic(fmt.Sprintf("unknown option: %s", opt))
		}
	}
	return t
}

type UUIDs []uuid.UUID

func (us *UUIDs) String() string {
	if us == nil {
		return "<nil>"
	}

	var s []string
	for _, u := range *us {
		s = append(s, u.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ", "))
}

func (us *UUIDs) Type() string {
	return "[UUID]"
}

func (us *UUIDs) Set(s string) error {
	u, err := uuid.FromString(s)
	*us = append(*us, u)
	return err
}

func (us *UUIDs) Slice() *[]uuid.UUID {
	if us == nil {
		return nil
	}
	return (*[]uuid.UUID)(us)
}
