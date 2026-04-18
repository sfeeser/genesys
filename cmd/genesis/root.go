package main

import (
	"github.com/spf13/cobra"
)

var (
	genomePath  string
	projectPath string
	specPath    string
	auditUnix   int64 // Set by main.go
)

var rootCmd = &cobra.Command{
	Use:   "genesis",
	Short: "The Code Genome Engine",
}

func init() {
	// PRECEDENCE: Flags > Env > Defaults
	rootCmd.PersistentFlags().StringVarP(&genomePath, "genome", "g", "genome.db", "Path to Registry")
	rootCmd.PersistentFlags().StringVarP(&projectPath, "path", "p", ".", "Target Project Root")
	rootCmd.PersistentFlags().StringVarP(&specPath, "spec", "s", "genesis.yaml", "Path to Specbook")
}

func Execute() {
	rootCmd.Execute()
}
