//go:build tools
// +build tools

// Package tools exists so development dependency versions can be tracked using
// the usual Go modules system.  These dependencies are all binaries, so it is
// impossible to actually build this package.
package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
