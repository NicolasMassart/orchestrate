package streams

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/notifier/service/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const notifyTransactionComponent = "use-cases.notify_transaction"

type notifyTransactionUseCase struct {
	db               store.EventStreamAgent
	messengerCli     messenger.Producer
	searchContractUC usecases.SearchContractUseCase
	decodeLogUC      usecases.DecodeEventLogUseCase
	notifierTopic    string
	logger           *log.Logger
}

func NewNotifyTransactionUseCase(
	db store.EventStreamAgent,
	messengerCli messenger.Producer,
	searchContractUC usecases.SearchContractUseCase,
	decodeLogUC usecases.DecodeEventLogUseCase,
	notifierTopic string,
) usecases.NotifyTransactionUseCase {
	return &notifyTransactionUseCase{
		db:               db,
		messengerCli:     messengerCli,
		searchContractUC: searchContractUC,
		decodeLogUC:      decodeLogUC,
		notifierTopic:    notifierTopic,
		logger:           log.NewLogger().SetComponent(notifyTransactionComponent),
	}
}

func (uc *notifyTransactionUseCase) Execute(ctx context.Context, job *entities.Job, errStr string, userInfo *multitenancy.UserInfo) error {
	ctx = log.WithFields(ctx, log.Field("id", job.ScheduleUUID))

	eventStream, err := uc.db.FindOneByTenantAndChain(ctx, job.TenantID, job.ChainUUID, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return errors.FromError(err).ExtendComponent(notifyTransactionComponent)
	}

	if eventStream == nil {
		return nil
	}

	if job.Status == entities.StatusFailed && job.InternalData.ParentJobUUID != "" { // We do not notify on failed children jobs
		return nil
	}

	logger := uc.logger.WithContext(ctx).WithField("event_stream", eventStream.Name).WithField("channel", eventStream.Channel)
	if job.Status == entities.StatusMined {
		if job.Receipt == nil {
			errMsg := "missing required receipt for notification"
			logger.Error(errMsg)
			return errors.InvalidParameterError(errMsg)
		}

		err = uc.attachContractData(ctx, job.Receipt)
		if err != nil {
			return errors.FromError(err).ExtendComponent(notifyTransactionComponent)
		}

		for idx, l := range job.Receipt.Logs {
			decodedLog, err2 := uc.decodeLogUC.Execute(ctx, job.ChainUUID, l)
			if err2 != nil {
				return errors.FromError(err2).ExtendComponent(notifyTransactionComponent)
			}
			if decodedLog != nil {
				job.Receipt.Logs[idx] = decodedLog
			}
		}
	}

	msg := &types.NotificationMessage{
		Type:        types.TransactionNotificationType,
		EventStream: eventStream,
		Job:         job,
		Error:       errStr,
	}
	err = uc.messengerCli.SendNotificationMessage(uc.notifierTopic, msg, job.PartitionKey(), userInfo)
	if err != nil {
		errMsg := "failed to send notification to notifier service"
		uc.logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg).ExtendComponent(notifyTransactionComponent)
	}

	logger.Debug("notification sent successfully")
	return nil
}

func (uc *notifyTransactionUseCase) attachContractData(ctx context.Context, receipt *ethereum.Receipt) error {
	var contractAddress *ethcommon.Address

	if receipt.ContractAddress != "" && receipt.ContractAddress != utils.ZeroAddressString {
		contractAddress = utils.ToPtr(ethcommon.HexToAddress(receipt.ContractAddress)).(*ethcommon.Address)
	} else {
		for _, l := range receipt.GetLogs() {
			if l.GetAddress() != "" {
				contractAddress = utils.ToPtr(ethcommon.HexToAddress(l.GetAddress())).(*ethcommon.Address)
				break
			}
		}
	}

	if contractAddress == nil {
		return nil
	}

	eventContract, err := uc.searchContractUC.Execute(ctx, nil, contractAddress)
	if err != nil {
		return err
	}

	if eventContract == nil {
		return nil
	}

	receipt.ContractName = eventContract.Name
	receipt.ContractTag = eventContract.Tag

	return nil
}
