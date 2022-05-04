package contracts

import (
	"context"
	"math/big"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-listener/store"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const registerDeployContractComponent = "tx-listener.use-case.tx-listener.register_deploy_contract"

type registerDeployedContractUseCase struct {
	chainState store.Chain
	client     sdk.OrchestrateClient
	ethClient  ethclient.MultiClient
	logger     *log.Logger
}

func RegisterDeployedContractUseCase(client sdk.OrchestrateClient,
	ethClient ethclient.MultiClient,
	chainState store.Chain,
	logger *log.Logger,
) usecases.RegisterDeployedContract {
	return &registerDeployedContractUseCase{
		client:     client,
		ethClient:  ethClient,
		chainState: chainState,
		logger:     logger.SetComponent(registerDeployContractComponent),
	}
}

func (uc *registerDeployedContractUseCase) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.
		WithField("chain", job.ChainUUID).
		WithField("job", job.UUID).
		WithField("contract_address", job.Receipt.GetContractAddr().String())
	logger.Debug("registering contract deployed")

	chainURL := uc.client.ChainProxyURL(job.ChainUUID)
	var code []byte
	var err error
	if job.Receipt.PrivacyGroupId != "" {
		// Fetch EEA deployed contract code
		code, err = uc.ethClient.PrivCodeAt(ctx, chainURL, ethcommon.HexToAddress(job.Receipt.ContractAddress),
			job.Receipt.PrivacyGroupId, new(big.Int).SetUint64(job.Receipt.BlockNumber))
	} else {
		code, err = uc.ethClient.CodeAt(ctx, chainURL, ethcommon.HexToAddress(job.Receipt.ContractAddress),
			new(big.Int).SetUint64(job.Receipt.BlockNumber))
	}
	if err != nil {
		errMsg := "failed to retrieve contract code"
		logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg)
	}

	chain, err := uc.chainState.Get(ctx, job.ChainUUID)
	if err != nil {
		logger.WithError(err).Error("failed to get chain for contract registration")
		return err
	}
	err = uc.client.SetContractAddressCodeHash(ctx, job.Receipt.ContractAddress, chain.ChainID.String(),
		&api.SetContractCodeHashRequest{
			CodeHash: crypto.Keccak256Hash(code).Bytes(),
		})
	if err != nil {
		errMsg := "failed to register contract"
		logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg)
	}

	logger.Info("contract has been registered successfully")
	return nil
}
