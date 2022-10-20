package main

import (
	"os"

	"github.com/bupt-narc/rinp/components/sidecar"
)

func main() {

	if err := sidecar.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
