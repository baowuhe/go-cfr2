package utils

import (
	"fmt"
	"os"
)

// ExitWithError prints an error message to stderr and exits the program with status code 1.
func ExitWithError(msg string) {
	fmt.Fprintf(os.Stderr, "× %s\n", msg)
	os.Exit(1)
}
