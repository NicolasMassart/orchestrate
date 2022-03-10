package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"

	qkm "github.com/consensys/quorum-key-manager/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAccountCmd() *cobra.Command {
	var postgresClient postgres.Client
	var qkmClient qkm.KeyManagerClient
	var storeName string

	accountCmd := &cobra.Command{
		Use:   "account",
		Short: "Account management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			vipr := viper.GetViper()

			// Set database connection
			postgresConfig := flags.NewPGConfig(vipr)
			postgresClient, err = gopg.New("orchestrate.accounts", postgresConfig)
			if err != nil {
				return err
			}

			// Init QKM client
			qkmConfig := flags.NewQKMConfig(vipr)
			qkmClient, err = http.New(qkmConfig)
			if err != nil {
				return err
			}

			storeName = qkmConfig.StoreName

			return nil
		},
	}

	// Postgres flags
	flags.PGFlags(accountCmd.Flags())
	flags.QKMFlags(accountCmd.Flags())

	importCmd := &cobra.Command{
		Use:   "import",
		Short: "import accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return importAccounts(cmd.Context(), postgresClient, qkmClient, storeName)
		},
	}
	accountCmd.AddCommand(importCmd)

	return accountCmd
}

func importAccounts(ctx context.Context, postgresClient postgres.Client, client qkm.KeyManagerClient, storeName string) error {
	log.Debug("Loading accounts from Vault...")

	accounts, err := client.ListEthAccounts(ctx, storeName, 0, 0)
	if err != nil {
		log.WithError(err).Errorf("could not get list of accounts")
		return err
	}

	var queryInsertItems []string
	for _, accountID := range accounts {
		acc, err2 := client.GetEthAccount(ctx, storeName, accountID)
		if err2 != nil {
			log.WithField("account_id", accountID).WithError(err2).Error("Could not get account")
			return err2
		}

		tenantIDs := strings.Split(acc.Tags[http.TagIDAllowedTenants], http.TagSeparatorAllowedTenants)
		for _, tenantID := range tenantIDs {
			queryInsertItems = append(queryInsertItems, fmt.Sprintf("('%s', '%s', '%s', '%s', '{\"source\": \"kv-v2\"}')",
				tenantID,
				acc.Address,
				acc.PublicKey,
				acc.CompressedPublicKey,
			))
		}
	}

	if len(queryInsertItems) > 0 {
		err = postgresClient.Exec("INSERT INTO accounts (tenant_id, address, public_key, compressed_public_key, attributes) VALUES " +
			strings.Join(queryInsertItems, ", ") + " on conflict do nothing")
		if err != nil {
			log.WithError(err).Error("Could not import accounts")
			return err
		}
	}

	log.WithField("accounts", len(queryInsertItems)).Info("accounts imported successfully")
	return nil
}
