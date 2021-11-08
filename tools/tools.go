//go:build tools
// +build tools

package tools // import "tools"

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/jcchavezs/porto/cmd/porto"
)
