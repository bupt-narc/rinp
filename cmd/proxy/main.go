package main

import (
	"os"

	"github.com/bupt-narc/rinp/proxy"
)

func main() {
	if err := proxy.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
