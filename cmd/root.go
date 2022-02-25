package cmd

import (
	"github.com/consensys/orchestrate/cmd/api"
	chainlistener "github.com/consensys/orchestrate/cmd/chain-listener"
	txsender "github.com/consensys/orchestrate/cmd/tx-sender"
	"github.com/spf13/cobra"
)

// NewCommand create root command
func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "orchestrate",
		TraverseChildren: true,
		SilenceUsage:     true,
	}

	// Add Run command
	rootCmd.AddCommand(txsender.NewRootCommand())
	rootCmd.AddCommand(chainlistener.NewRootCommand())
	rootCmd.AddCommand(api.NewRootCommand())

	return rootCmd
}
