package main

import (
	"github.com/user/lc/cmd/lc"
)

// 版本信息，由 Makefile 在构建时注入
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	cmd.SetVersionInfo(Version, Commit, BuildTime)
	cmd.Execute()
}
