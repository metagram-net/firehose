package clio

import (
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
)

type Verbosity int

const (
	SilentLevel Verbosity = iota - 1
	DefaultLevel
	InfoLevel
	DebugLevel
)

type IO struct {
	stdout io.Writer
	stderr io.Writer

	verbosity Verbosity
}

func New(opts ...Option) *IO {
	io := &IO{
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		verbosity: DefaultLevel,
	}
	for _, opt := range opts {
		opt(io)
	}
	return io
}

type Option func(*IO)

func WithOutput(w io.Writer) Option {
	return func(io *IO) { io.stdout = w }
}

func WithError(w io.Writer) Option {
	return func(io *IO) { io.stderr = w }
}

func WithVerbosity(v Verbosity) Option {
	return func(io *IO) { io.verbosity = v }
}

func (io *IO) SetVerbosity(v Verbosity) {
	io.verbosity = v
}

func (io IO) Errorf(format string, args ...interface{}) {
	io.logf(DefaultLevel, format, args...)
}

func (io IO) Infof(format string, args ...interface{}) {
	io.logf(InfoLevel, format, args...)
}

func (io IO) Debugf(format string, args ...interface{}) {
	io.logf(DebugLevel, format, args...)
}

func (io IO) Logf(format string, args ...interface{}) {
	io.logf(DefaultLevel, format, args...)
}

func (io IO) logf(level Verbosity, format string, args ...interface{}) {
	if io.verbosity < level {
		return
	}
	fmt.Fprintf(io.stderr, format+"\n", args...)
}

func (io IO) Printf(format string, args ...interface{}) {
	if io.verbosity < DefaultLevel {
		return
	}
	fmt.Fprintf(io.stdout, format+"\n", args...)
}

func (io IO) LogTable(cols []string, rows [][]string) {
	if io.verbosity < DefaultLevel {
		return
	}

	t := tablewriter.NewWriter(io.stderr)
	t.SetAutoFormatHeaders(false)
	t.SetHeader(cols)
	t.AppendBulk(rows)
	t.Render()
}
