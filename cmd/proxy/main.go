package main

import (
	"os"

	"github.com/bupt-narc/rinp/pkg/cmd/proxy"
)

func main() {
	if err := proxy.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
