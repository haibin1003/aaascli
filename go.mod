module github.com/haibin1003/aaascli

go 1.25.0

require (
	github.com/spf13/cobra v1.8.0
	golang.org/x/sys v0.43.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/text v0.21.0
)

replace golang.org/x/text => ./third_party/golang.org/x/text
