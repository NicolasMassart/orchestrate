package sender

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
	utils2 "github.com/consensys/orchestrate/src/tx-sender/tx-sender/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const sendGoQuorumPrivateTxComponent = "use-cases.send-go-quorum-private-tx"

type sendGoQuorumPrivateTxUseCase struct {
	ec               ethclient.QuorumTransactionSender
	chainRegistryURL string
	jobClient        client.JobClient
	crafter          usecases.CraftTransactionUseCase
	logger           *log.Logger
}

func NewSendGoQuorumPrivateTxUseCase(ec ethclient.QuorumTransactionSender, crafter usecases.CraftTransactionUseCase,
	jobClient client.JobClient, chainRegistryURL string) usecases.SendGoQuorumPrivateTxUseCase {
	return &sendGoQuorumPrivateTxUseCase{
		ec:               ec,
		chainRegistryURL: chainRegistryURL,
		jobClient:        jobClient,
		crafter:          crafter,
		logger:           log.NewLogger().SetComponent(sendGoQuorumPrivateTxComponent),
	}
}

func (uc *sendGoQuorumPrivateTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	ctx = log.With(log.WithFields(
		ctx,
		log.Field("job", job.UUID),
		log.Field("schedule_uuid", job.ScheduleUUID),
	), uc.logger)
	logger := uc.logger.WithContext(ctx)
	logger.Debug("processing tessera private transaction job")

	job.Transaction.Nonce = utils.ToPtr(uint64(0)).(*uint64)
	err := uc.crafter.Execute(ctx, job)
	if err != nil {
		return errors.FromError(err).ExtendComponent(sendGoQuorumMarkingTxComponent)
	}

	job.Transaction.EnclaveKey, err = uc.sendTx(ctx, job)
	if err != nil {
		return errors.FromError(err).ExtendComponent(sendGoQuorumPrivateTxComponent)
	}

	err = utils2.UpdateJobStatus(ctx, uc.jobClient, job, entities.StatusStored, "", job.Transaction)
	if err != nil {
		return errors.FromError(err).ExtendComponent(sendGoQuorumPrivateTxComponent)
	}

	logger.Info("tessera private job was sent successfully")
	return nil
}

func (uc *sendGoQuorumPrivateTxUseCase) sendTx(ctx context.Context, job *entities.Job) (hexutil.Bytes, error) {
	logger := uc.logger.WithContext(ctx)
	proxyTessera := utils.GetProxyTesseraURL(uc.chainRegistryURL, job.ChainUUID)

	enclaveKey, err := uc.ec.StoreRaw(ctx, proxyTessera, job.Transaction.Data, job.Transaction.PrivateFrom)
	if err != nil {
		errMsg := "cannot send tessera private transaction"
		logger.WithError(err).Error(errMsg)
		return nil, err
	}

	return enclaveKey, nil
}
