package api

import (
	"os"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/src/api"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdErr error

func newRunCommand() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run application",
		RunE:  run,
		PreRun: func(cmd *cobra.Command, args []string) {
			utils.PreRunBindFlags(viper.GetViper(), cmd.Flags(), "api")
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			if err := errors.CombineErrors(cmdErr, cmd.Context().Err()); err != nil {
				os.Exit(1)
			}
		},
	}

	flags.NewAPIFlags(runCmd.Flags())

	return runCmd
}

func run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	apiApp, err := api.New(ctx, flags.NewAPIConfig(viper.GetViper()))
	if err != nil {
		return err
	}

	return apiApp.Run(ctx)
}
