package main

import (
	channelcheck "github.com/asymmetric-research/channel_linter"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(channelcheck.Analyzer)
}
