package main

import (
	"os"

	"github.com/bupt-narc/rinp/sidecar"
)

func main() {

	if err := sidecar.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
