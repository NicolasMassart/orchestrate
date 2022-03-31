package main

import (
	"context"
	"os"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/spf13/viper"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/tests/service/e2e"
	"github.com/spf13/cobra"
)

func NewRunE2ECommand() *cobra.Command {
	runCmd := &cobra.Command{
		Use:          "e2e",
		Short:        "Run e2e test",
		RunE:         runE2E,
		SilenceUsage: true,
	}

	e2e.Flags(runCmd.Flags())
	return runCmd
}

func runE2E(cmd *cobra.Command, _ []string) error {
	logger := log.NewLogger().SetComponent("e2e.cucumber")
	ctx, cancel := context.WithCancel(cmd.Context())
	ctx = log.With(ctx, logger)

	// Process signals
	sig := utils.NewSignalListener(func(signal os.Signal) {
		cancel()
	})
	defer sig.Close()

	cfg, err := e2e.NewConfig(viper.GetViper())
	if err != nil {
		logger.WithError(err).Error("failed to load settings")
		return err
	}

	var gerr error
	if err := e2e.Start(ctx, cfg); err != nil {
		logger.WithError(err).Error("did not complete successfully")
		gerr = errors.CombineErrors(gerr, err)
	}

	if err := e2e.Stop(ctx); err != nil {
		logger.WithError(err).Error("did not shutdown properly")
		gerr = errors.CombineErrors(gerr, err)
	}

	return gerr
}
