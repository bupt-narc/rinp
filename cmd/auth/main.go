package main

import (
	"fmt"
	"os"

	"github.com/bupt-narc/rinp/components/auth"
)

func main() {
	if err := auth.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}
