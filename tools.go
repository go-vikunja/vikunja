// +build tools

package tools

// This file is needed for go mod to recognize the tools we use.

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/fzipp/gocyclo"
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/gordonklaus/ineffassign"
	_ "github.com/karalabe/xgo"
	_ "golang.org/x/lint/golint"
)
