// Path: genesis/cmd/genesis/root.go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// Execute is the entry point called by main.go
func Execute() {
	// CHAPTER 14.3: Root-level Context Ownership
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		code := 1 // Default: Panic/General Failure

		// CHAPTER 14.4: DETERMINISTIC EXIT CODES (Typed Switch)
		switch {
		case errors.Is(err, ErrDeterminantMissing):
			code = 2
		case errors.Is(err, ErrAccessDenied):
			code = 126
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			code = 130
		}

		fmt.Fprintf(os.Stderr, "GENESIS ERROR: %v\n", err)
		os.Exit(code)
	}
}
