package sender

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/nonce"
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
	utils2 "github.com/consensys/orchestrate/src/tx-sender/tx-sender/utils"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const sendGoQuorumMarkingTxComponent = "use-cases.send-go-quorum-marking-tx"

type sendGoQuorumMarkingTxUseCase struct {
	signTx           usecases.SignETHTransactionUseCase
	crafter          usecases.CraftTransactionUseCase
	nonceChecker     nonce.Manager
	jobClient        client.JobClient
	ec               ethclient.QuorumTransactionSender
	chainRegistryURL string
	logger           *log.Logger
}

func NewSendGoQuorumMarkingTxUseCase(signTx usecases.SignGoQuorumPrivateTransactionUseCase,
	crafter usecases.CraftTransactionUseCase,
	ec ethclient.QuorumTransactionSender,
	jobClient client.JobClient,
	chainRegistryURL string,
	nonceChecker nonce.Manager,
) usecases.SendGoQuorumMarkingTxUseCase {
	return &sendGoQuorumMarkingTxUseCase{
		signTx:           signTx,
		nonceChecker:     nonceChecker,
		ec:               ec,
		jobClient:        jobClient,
		chainRegistryURL: chainRegistryURL,
		crafter:          crafter,
		logger:           log.NewLogger().SetComponent(sendGoQuorumMarkingTxComponent),
	}
}

func (uc *sendGoQuorumMarkingTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	ctx = log.With(log.WithFields(
		ctx,
		log.Field("job", job.UUID),
		log.Field("tenant_id", job.TenantID),
		log.Field("owner_id", job.OwnerID),
		log.Field("schedule_uuid", job.ScheduleUUID),
	), uc.logger)
	logger := uc.logger.WithContext(ctx)
	logger.Debug("processing tessera marking transaction job")

	err := uc.crafter.Execute(ctx, job)
	if err != nil {
		return errors.FromError(err).ExtendComponent(sendGoQuorumMarkingTxComponent)
	}

	// In case of job resending we don't need to sign again
	if job.InternalData.ParentJobUUID == job.UUID || job.Status == entities.StatusPending || job.Status == entities.StatusResending {
		err = utils2.UpdateJobStatus(ctx, uc.jobClient, job, entities.StatusResending, "", job.Transaction)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendETHTxComponent)
		}
	} else {
		job.Transaction.Raw, job.Transaction.Hash, err = uc.signTx.Execute(ctx, job)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendGoQuorumMarkingTxComponent)
		}

		err = utils2.UpdateJobStatus(ctx, uc.jobClient, job, entities.StatusPending, "", job.Transaction)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendGoQuorumMarkingTxComponent)
		}
	}

	txHash, err := uc.sendTx(ctx, job)
	if err != nil {
		if err2 := uc.nonceChecker.CleanNonce(ctx, job, err); err2 != nil {
			return errors.FromError(err2).ExtendComponent(sendGoQuorumMarkingTxComponent)
		}
		return errors.FromError(err).ExtendComponent(sendGoQuorumMarkingTxComponent)
	}

	err = uc.nonceChecker.IncrementNonce(ctx, job)
	if err != nil {
		return err
	}

	if txHash.String() != job.Transaction.Hash.String() {
		warnMessage := fmt.Sprintf("expected transaction hash %s, but got %s. Overriding", job.Transaction.Hash, txHash)
		job.Transaction.Hash = txHash
		err = utils2.UpdateJobStatus(ctx, uc.jobClient, job, entities.StatusWarning, warnMessage, job.Transaction)
		if err != nil {
			return errors.FromError(err).ExtendComponent(sendGoQuorumMarkingTxComponent)
		}
	}

	logger.Info("tessera marking transaction job was sent successfully")
	return nil
}

func (uc *sendGoQuorumMarkingTxUseCase) sendTx(ctx context.Context, job *entities.Job) (*ethcommon.Hash, error) {
	proxyURL := client.GetProxyURL(uc.chainRegistryURL, job.ChainUUID)
	txHash, err := uc.ec.SendQuorumRawPrivateTransaction(ctx, proxyURL, job.Transaction.Raw,
		job.Transaction.PrivateFor, job.Transaction.MandatoryFor,
		int(job.Transaction.PrivacyFlag))
	if err != nil {
		uc.logger.WithContext(ctx).WithError(err).Error("cannot send tessera marking transaction")
		return nil, err
	}

	return &txHash, nil
}
