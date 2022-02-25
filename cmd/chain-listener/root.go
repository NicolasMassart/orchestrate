package txlistener

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "chain-listener",
		Short: "Run chain-listener",
	}

	rootCmd.AddCommand(newRunCommand())

	return rootCmd
}
