package main

import (
	"os"

	"github.com/bupt-narc/rinp/client"
)

func main() {
	if err := client.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
