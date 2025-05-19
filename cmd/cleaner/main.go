package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gi8lino/cleaner/internal/app"
)

var (
	version   = "dev"  // set via build flag
	gitCommit = "none" // set via build flag
)

// main sets up the application context and runs the proxy.
func main() {
	ctx := context.Background()

	if err := app.Run(ctx, version, gitCommit, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
