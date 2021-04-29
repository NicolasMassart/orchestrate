package sender

import (
	"context"
	"fmt"

	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/pkg/sdk/client"
	"github.com/ConsenSys/orchestrate/pkg/toolkit/app/log"
	"github.com/ConsenSys/orchestrate/pkg/toolkit/ethclient"
	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	"github.com/ConsenSys/orchestrate/pkg/utils"
	"github.com/ConsenSys/orchestrate/services/tx-sender/tx-sender/nonce"
	usecases "github.com/ConsenSys/orchestrate/services/tx-sender/tx-sender/use-cases"
	utils2 "github.com/ConsenSys/orchestrate/services/tx-sender/tx-sender/utils"
)

const sendETHTxComponent = "use-cases.send-eth-tx"

type sendETHTxUseCase struct {
	signTx           usecases.SignETHTransactionUseCase
	crafter          usecases.CraftTransactionUseCase
	nonceChecker     nonce.Manager
	ec               ethclient.TransactionSender
	chainRegistryURL string
	jobClient        client.JobClient
	logger           *log.Logger
}

func NewSendEthTxUseCase(signTx usecases.SignETHTransactionUseCase, crafter usecases.CraftTransactionUseCase,
	ec ethclient.TransactionSender, jobClient client.JobClient, chainRegistryURL string,
	nonceChecker nonce.Manager) usecases.SendETHTxUseCase {
	return &sendETHTxUseCase{
		jobClient:        jobClient,
		ec:               ec,
		chainRegistryURL: chainRegistryURL,
		signTx:           signTx,
		nonceChecker:     nonceChecker,
		crafter:          crafter,
		logger:           log.NewLogger().SetComponent(sendETHTxComponent),
	}
}

func (uc *sendETHTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	ctx = log.With(log.WithFields(
		ctx,
		log.Field("job", job.UUID),
		log.Field("tenant_id", job.TenantID),
		log.Field("schedule_uuid", job.ScheduleUUID),
	), uc.logger)
	logger := uc.logger.WithContext(ctx)
	logger.Debug("processing ethereum transaction job")

	err := uc.crafter.Execute(ctx, job)
	if err != nil {
		return errors.FromError(err).ExtendComponent(sendETHTxComponent)
	}

	// In case of job resending we don't need to sign again
	if job.InternalData.ParentJobUUID == job.UUID {
		err = utils2.UpdateJobStatus(ctx, uc.jobClient, job.UUID, entities.StatusResending, "", job.Transaction)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendETHTxComponent)
		}
	} else {
		job.Transaction.Raw, job.Transaction.Hash, err = uc.signTx.Execute(ctx, job)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendETHTxComponent)
		}

		err = utils2.UpdateJobStatus(ctx, uc.jobClient, job.UUID, entities.StatusPending, "", job.Transaction)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendETHTxComponent)
		}
	}

	txHash, err := uc.sendTx(ctx, job)
	if err != nil {
		if err2 := uc.nonceChecker.CleanNonce(ctx, job, err); err2 != nil {
			return errors.FromError(err2).ExtendComponent(sendETHTxComponent)
		}
		return errors.FromError(err).ExtendComponent(sendETHTxComponent)
	}

	if err = uc.nonceChecker.IncrementNonce(ctx, job); err != nil {
		return errors.FromError(err).ExtendComponent(sendETHTxComponent)
	}

	if txHash != job.Transaction.Hash {
		warnMessage := fmt.Sprintf("expected transaction hash %s, but got %s. overriding", job.Transaction.Hash, txHash)
		job.Transaction.Hash = txHash
		err = utils2.UpdateJobStatus(ctx, uc.jobClient, job.UUID, entities.StatusWarning, warnMessage, job.Transaction)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendETHTxComponent)
		}
	}

	logger.Info("ethereum transaction job was sent successfully")
	return nil
}

func (uc *sendETHTxUseCase) sendTx(ctx context.Context, job *entities.Job) (string, error) {
	proxyURL := utils.GetProxyURL(uc.chainRegistryURL, job.ChainUUID)
	txHash, err := uc.ec.SendRawTransaction(ctx, proxyURL, job.Transaction.Raw)
	if err != nil {
		uc.logger.WithContext(ctx).WithError(err).Error("cannot send raw ethereum transaction")
		return "", err
	}

	return txHash.String(), nil
}
