// Path: genesis/cmd/genesis/root.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	genomePath  string
	projectPath string
	specPath    string
	auditUnix   int64
)

var rootCmd = &cobra.Command{
	Use:   "genesis",
	Short: "The Code Genome Engine",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// CHAPTER 15: CLOCK INJECTION
		auditUnix = time.Now().UTC().Unix()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&genomePath, "genome", "g", "genome.db", "Path to Registry")
	rootCmd.PersistentFlags().StringVarP(&projectPath, "path", "p", ".", "Target Project Root")
	rootCmd.PersistentFlags().StringVarP(&specPath, "spec", "s", "genesis.yaml", "Path to Specbook")
}

func Execute() {
	// CHAPTER 14.3: CONTEXT OWNERSHIP (Root Level)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
