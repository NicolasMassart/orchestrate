package txsender

import (
	"os"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
	txsender "github.com/consensys/orchestrate/src/tx-sender"
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
			utils.PreRunBindFlags(viper.GetViper(), cmd.Flags(), "tx-sender")
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			if err := errors.CombineErrors(cmdErr, cmd.Context().Err()); err != nil {
				os.Exit(1)
			}
		},
	}

	// Register KeyStore flags
	flags.TxSenderFlags(runCmd.Flags())

	return runCmd
}

func run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	cfg := flags.NewTxSenderConfig(viper.GetViper())

	app, err := txsender.New(ctx, cfg)
	if err != nil {
		return errors.CombineErrors(cmdErr, err)
	}

	err = app.Run(ctx)
	if err != nil {
		return errors.CombineErrors(cmdErr, err)
	}

	return nil
}
