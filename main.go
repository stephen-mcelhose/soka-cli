package main

import (
	"fmt"
	"os"

	"github.com/stephen-mcelhose/soka-cli/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "soka: %v\n", err)
		os.Exit(1)
	}
}
