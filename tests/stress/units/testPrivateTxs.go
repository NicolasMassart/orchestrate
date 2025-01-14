package units

import (
	"context"
	"encoding/json"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/tests/pkg/trackers"
	"github.com/consensys/orchestrate/tests/stress/assets"
)

func BatchPrivateTxsTest(ctx context.Context, cfg *WorkloadConfig, client sdk.OrchestrateClient,
	consumerTracker *trackers.NotifierConsumerTracker) error {
	logger := log.WithContext(ctx).SetComponent("stress-test.private-txs")

	account := cfg.accounts[utils.RandInt(len(cfg.accounts))]
	contractName := cfg.artifacts[utils.RandInt(len(cfg.artifacts))]
	chain := cfg.chains[utils.RandInt(len(cfg.chains))]
	privacyGroup := cfg.privacyGroups[utils.RandInt(len(cfg.privacyGroups))]
	privateFrom := chain.PrivNodeAddress[utils.RandInt(len(chain.PrivNodeAddress))]
	idempotency := utils.RandString(30)

	req := &api.DeployContractRequest{
		ChainName: chain.Name,
		Params: api.DeployContractParams{
			From:         &account,
			ContractName: contractName,
			Args:         constructorArgs(contractName),
			PrivateFrom:  privateFrom,
			Protocol:     entities.EEAChainType,
		},
		Labels: map[string]string{
			"id": idempotency,
		},
	}

	usePrivacyGroup := canUsePrivacyGroup(chain.PrivNodeAddress, &privacyGroup)
	if usePrivacyGroup {
		req.Params.PrivacyGroupID = privacyGroup.ID
	} else {
		size := len(privacyGroup.Nodes)
		req.Params.PrivateFor = privacyGroup.Nodes[0 : size-1]
	}

	sReq, _ := json.Marshal(req)
	logger = logger.WithField("chain", req.ChainName).WithField("idem", idempotency)
	logger.Debug("sending private tx to deploy contract")

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

func canUsePrivacyGroup(chainPrivNodes []string, pGroup *assets.PrivacyGroup) bool {
	for _, cAddr := range chainPrivNodes {
		for _, gAddr := range pGroup.Nodes {
			if cAddr == gAddr {
				return true
			}
		}
	}

	return false
}
