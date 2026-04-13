package main

import (
	"github.com/haibin1003/aaascli/cmd/sdp"
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
