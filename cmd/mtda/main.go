package main

import (
	"os"

	"github.com/bupt-narc/mtda/pkg/cmd"
)

func main() {
	if err := cmd.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
