package main

import (
	"fmt"
	"os"

	"github.com/isaacgarza/dev-stack/internal/cli"
)

func main() {
	// Try to use the factory-based CLI first
	if err := cli.ExecuteFactory(); err != nil {
		// If factory fails, fall back to basic CLI silently
		if fallbackErr := cli.Execute(); fallbackErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", fallbackErr)
			os.Exit(1)
		}
	}
}
