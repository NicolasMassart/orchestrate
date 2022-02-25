package txlistener

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/types"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
)

const updateChainHeadComponent = "chain-listener.use-case.tx-listener.update-chain-head"

type updateChainHead struct {
	client orchestrateclient.OrchestrateClient
	logger *log.Logger
}

func UpdateChainHeadUseCase(client orchestrateclient.OrchestrateClient, logger *log.Logger) usecases.UpdateChainHead {
	return &updateChainHead{
		client: client,
		logger: logger.SetComponent(updateChainHeadComponent),
	}
}

func (uc *updateChainHead) Execute(ctx context.Context, chainUUID string, nextChainHead uint64) error {
	logger := uc.logger.WithField("head", nextChainHead)
	_, err := uc.client.UpdateChain(ctx, chainUUID, &types.UpdateChainRequest{
		Listener: &types.UpdateListenerRequest{
			CurrentBlock: nextChainHead,
		},
	})

	if err != nil {
		errMsg := "failed to update chain head"
		logger.WithError(err).Error(errMsg)
		return errors.FromError(err).SetMessage(errMsg)
	}

	uc.logger.WithField("head", nextChainHead).Debug("chain head updated")
	return nil
}
