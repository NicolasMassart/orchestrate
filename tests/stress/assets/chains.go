package assets

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/sdk"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/tests/config"
	pkgutils "github.com/consensys/orchestrate/tests/pkg/utils"
)

var chainsCtxKey ctxKey = "chains"

type RegisteredChainData struct {
	UUID            string
	Name            string
	ProxyURL        string
	PrivNodeAddress []string
	EventStreamUUID string
}

func RegisterNewChainWithEventStream(ctx context.Context, client sdk.OrchestrateClient, ec ethclient.ChainSyncReader,
	proxyHost, chainName string, chainData *config.ChainDataCfg, kafkaTopic string,
) (context.Context, string, error) {
	logger := log.FromContext(ctx).WithField("name", chainName).WithField("urls", chainData.URLs)
	logger.WithContext(ctx).Debug("registering new chain")

	c, err := client.RegisterChain(ctx, &api.RegisterChainRequest{
		Name: chainName,
		URLs: chainData.URLs,
	})
	if err != nil {
		logger.WithError(err).Error("failed to register chain")
		return nil, "", err
	}

	err = pkgutils.WaitForProxyWithHost(ctx, proxyHost, c.UUID, ec)
	if err != nil {
		logger.WithError(err).Error("failed to wait for chain proxy")
		return nil, "", err
	}

	es, err := client.CreateKafkaEventStream(ctx, &api.CreateKafkaEventStreamRequest{
		Name:  "chain-" + c.UUID, // Using same reference for future deletion
		Topic: kafkaTopic,
		Chain: chainName,
	})
	if err != nil {
		logger.WithError(err).Error("failed to create event stream chain")
		return nil, "", err
	}

	logger.WithField("chain", c.UUID).Info("new chain has been registered")
	return contextWithChains(ctx, append(ContextChains(ctx),
		RegisteredChainData{
			UUID:            c.UUID,
			ProxyURL:        orchestrateclient.GetProxyURL(proxyHost, c.UUID),
			Name:            chainName,
			PrivNodeAddress: chainData.PrivateAddress,
			EventStreamUUID: es.UUID,
		}),
	), c.UUID, nil
}

func DeregisterChainAndEventStreams(ctx context.Context, client sdk.OrchestrateClient, chain *RegisteredChainData) error {
	logger := log.FromContext(ctx).WithField("uuid", chain.UUID).WithField("name", chain.Name)
	logger.WithContext(ctx).Debug("deleting chain")

	err := client.DeleteChain(ctx, chain.UUID)
	if err != nil {
		errMsg := "failed to delete chain"
		logger.WithError(err).Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	err = client.DeleteEventStream(ctx, chain.EventStreamUUID)
	if err != nil {
		errMsg := "failed to delete event stream"
		logger.WithError(err).Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	logger.Info("chain has been deleted successfully")
	return nil
}

func contextWithChains(ctx context.Context, chains []RegisteredChainData) context.Context {
	return context.WithValue(ctx, chainsCtxKey, chains)
}

func ContextChains(ctx context.Context) []RegisteredChainData {
	if v, ok := ctx.Value(chainsCtxKey).([]RegisteredChainData); ok {
		return v
	}

	return make([]RegisteredChainData, 0)
}
