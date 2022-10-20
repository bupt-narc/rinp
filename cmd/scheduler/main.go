package main

import (
	"os"

	"github.com/bupt-narc/rinp/components/scheduler"
)

func main() {
	if err := scheduler.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
