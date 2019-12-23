package provider

import (
	"context"

	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/tx-listener/dynamic"
)

// Provider defines methods of a provider.
type Provider interface {
	// Run starts the provider to provide configuration to the tx-listener
	// Canceling ctx stops the provider
	// Once context is canceled Run should not send any message into the configuration input
	Run(ctx context.Context, configInput chan<- *dynamic.Message) error
}