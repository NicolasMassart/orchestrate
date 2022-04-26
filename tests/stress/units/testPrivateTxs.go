package units

import (
	"context"
	"encoding/json"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	testutils2 "github.com/consensys/orchestrate/src/infra/notifier/kafka/testutils"
	"github.com/consensys/orchestrate/tests/stress/assets"
)

func BatchPrivateTxsTest(ctx context.Context, cfg *WorkloadConfig, client orchestrateclient.OrchestrateClient,
	consumerTracker *testutils2.NotifierConsumerTracker) error {
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

	_, err = consumerTracker.WaitForTxMinedNotification(ctx, tx.UUID, cfg.kafkaTopic, cfg.waitForEnvelopeTimeout)
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
