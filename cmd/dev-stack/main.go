package main

import (
	"fmt"
	"os"

	"github.com/isaacgarza/dev-stack/internal/cli"
)

func main() {
	if err := cli.ExecuteFactory(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
