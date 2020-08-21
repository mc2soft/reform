// +build tools

package tools // import "github.com/mc2soft/reform/tools"

import (
	_ "github.com/AlekSi/gocoverutil"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/quasilyte/go-consistent"
	_ "github.com/reviewdog/reviewdog/cmd/reviewdog"
	_ "golang.org/x/tools/cmd/goimports"
)
