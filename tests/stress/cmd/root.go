package main

import (
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	command := &cobra.Command{
		Use:              "run",
		TraverseChildren: true,
		SilenceUsage:     true,
	}

	// Set pkglog flags
	log.Flags(command.Flags())

	command.AddCommand(NewRunStressTestCommand())

	// Register Multi-Tenancy flags
	multitenancy.Flags(command.Flags())
	auth.Flags(command.Flags())

	if err := command.Execute(); err != nil {
		logrus.WithError(err).Fatalf("test: execution failed")
	}

	logrus.Infof("test: execution completed")
}