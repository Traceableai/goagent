//go:build tools
// +build tools

package tools // import "tools"

import (
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "github.com/jcchavezs/porto/cmd/porto"
)
