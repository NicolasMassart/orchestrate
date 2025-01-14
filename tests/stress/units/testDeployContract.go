package units

import (
	"context"
	"encoding/json"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/tests/pkg/trackers"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
)

func BatchDeployContractTest(ctx context.Context, cfg *WorkloadConfig, client sdk.OrchestrateClient,
	consumerTracker *trackers.NotifierConsumerTracker) error {
	logger := log.WithContext(ctx).SetComponent("stress-test.deploy-contract")
	nAccount := utils.RandInt(len(cfg.accounts))
	nArtifact := utils.RandInt(len(cfg.artifacts))
	nChain := utils.RandInt(len(cfg.chains))
	idempotency := utils.RandString(30)

	req := &api.DeployContractRequest{
		ChainName: cfg.chains[nChain].Name,
		Params: api.DeployContractParams{
			From:         &cfg.accounts[nAccount],
			ContractName: cfg.artifacts[nArtifact],
			Args:         constructorArgs(cfg.artifacts[nArtifact]),
		},
		Labels: map[string]string{
			"id": idempotency,
		},
	}
	sReq, _ := json.Marshal(req)

	logger = logger.WithField("chain", req.ChainName).WithField("idem", idempotency)
	tx, err := client.SendDeployTransaction(ctx, req)

	if err != nil {
		if !errors.IsConnectionError(err) {
			logger = logger.WithField("req", string(sReq))
		}
		logger.WithError(err).Error("failed to send transaction")
		return err
	}

	_, err = consumerTracker.WaitForMinedTransaction(ctx, tx.UUID, cfg.waitForEnvelopeTimeout)
	if err != nil {
		if !errors.IsConnectionError(err) {
			logger = logger.WithField("req", string(sReq))
		}
		logger.WithError(err).Error("failed to fetch envelope")
		return err
	}

	return nil
}
