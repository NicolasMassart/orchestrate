package utils

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/infra/ethclient"
)

func WaitForProxyWithHost(ctx context.Context, proxyHost, chainUUID string, ec ethclient.ChainSyncReader) error {
	chainProxyURL := client.GetProxyURL(proxyHost, chainUUID)
	return backoff.RetryNotify(
		func() error {
			_, err2 := ec.Network(ctx, chainProxyURL)
			return err2
		},
		backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 5),
		func(err error, duration time.Duration) {
			log.FromContext(ctx).WithField("chain", chainUUID).WithError(err).Debug("chain proxy is still not ready")
		},
	)
}

func WaitForProxy(ctx context.Context, chainProxyURL string, ec ethclient.ChainSyncReader) error {
	return backoff.RetryNotify(
		func() error {
			_, err2 := ec.Network(ctx, chainProxyURL)
			return err2
		},
		backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 5),
		func(err error, duration time.Duration) {
			log.FromContext(ctx).WithError(err).Debug("chain proxy is still not ready")
		},
	)
}

func RegisterChainAndWaitForProxy(ctx context.Context, sdkClient client.OrchestrateClient, ec ethclient.ChainSyncReader, req *types.RegisterChainRequest) (*types.ChainResponse, error) {
	chainRes, err := sdkClient.RegisterChain(ctx, req)
	if err != nil {
		return nil, err
	}
	chainProxyURL := sdkClient.ChainProxyURL(chainRes.UUID)
	err = WaitForProxy(ctx, chainProxyURL, ec)
	return chainRes, err
}
