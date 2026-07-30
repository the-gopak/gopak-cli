package main

import (
	"os"

	"github.com/the-gopak/gopak-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
