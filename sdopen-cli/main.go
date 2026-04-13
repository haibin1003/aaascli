package main

import (
	"github.com/yourname/sdopen-cli/cmd/sdp"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	cmd.SetVersionInfo(Version, Commit, BuildTime)
	cmd.Execute()
}
