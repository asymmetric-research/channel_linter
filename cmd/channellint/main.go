package main

import (
	"channel_linter/channelcheck"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(channelcheck.Analyzer)
}
