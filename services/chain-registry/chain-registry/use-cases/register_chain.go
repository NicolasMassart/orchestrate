package usecases

import (
	"context"

	"github.com/containous/traefik/v2/pkg/log"
	"github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/ethereum/ethclient"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/chain-registry/utils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store/models"
)

const registerChainComponent = "use-cases.register-chain"

type RegisterChain interface {
	Execute(ctx context.Context, chain *models.Chain) error
}

// RegisterContract is a use case to register a new contract
type registerChain struct {
	chainAgent store.ChainAgent
	ethClient  ethclient.ChainLedgerReader
}

// NewGetCatalog creates a new GetCatalog
func NewRegisterChain(chainAgent store.ChainAgent, ec ethclient.ChainLedgerReader) RegisterChain {
	return &registerChain{
		chainAgent: chainAgent,
		ethClient:  ec,
	}
}

func (uc *registerChain) Execute(ctx context.Context, chain *models.Chain) error {
	logger := log.FromContext(ctx)
	// In case of not staring block, we use latest
	if chain.ListenerStartingBlock == nil {
		head, err := utils.GetChainTip(ctx, uc.ethClient, chain.URLs)
		if err != nil {
			return errors.FromError(err).ExtendComponent(registerChainComponent)
		}

		chain.ListenerStartingBlock = &head
	}

	chain.SetDefault()
	err := uc.chainAgent.RegisterChain(ctx, chain)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"chainName": chain.Name,
		"chainUUID": chain.UUID,
		"tenantID":  chain.TenantID,
	}).Infof("registered a new chain")

	return nil
}
