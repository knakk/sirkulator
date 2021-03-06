//go:build tools
// +build tools

package sirkulator

// These imports ensure build tools are included in Go modules.
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
import (
	_ "github.com/benbjohnson/ego"
	_ "golang.org/x/text/cmd/gotext"
)
