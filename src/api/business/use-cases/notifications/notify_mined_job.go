package notifications

import (
	"context"

	"github.com/Shopify/sarama"
	encoding "github.com/consensys/orchestrate/pkg/encoding/proto"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	pkgsarama "github.com/consensys/orchestrate/src/infra/broker/sarama"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const notifyMinedJobComponent = "use-cases.notify-mined-job"

type notifyMinedJobUseCase struct {
	db               store.DB
	kafkaProducer    sarama.SyncProducer
	topicsCfg        *pkgsarama.KafkaTopicConfig
	searchContractUC usecases.SearchContractUseCase
	decodeLogUC      usecases.DecodeEventLogUseCase
	logger           *log.Logger
}

func NewNotifyMinedJobUseCase(
	db store.DB,
	searchContractUC usecases.SearchContractUseCase,
	decodeLogUC usecases.DecodeEventLogUseCase,
	kafkaProducer sarama.SyncProducer,
	topicsCfg *pkgsarama.KafkaTopicConfig,
) usecases.NotifyMinedJob {
	return &notifyMinedJobUseCase{
		db:               db,
		kafkaProducer:    kafkaProducer,
		topicsCfg:        topicsCfg,
		searchContractUC: searchContractUC,
		decodeLogUC:      decodeLogUC,
		logger:           log.NewLogger().SetComponent(notifyMinedJobComponent),
	}
}

func (uc *notifyMinedJobUseCase) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("job", job.UUID)

	if job.Receipt == nil {
		errMsg := "missing transaction receipt"
		uc.logger.Errorf(errMsg)
		return errors.InvalidFormatError(errMsg)
	}

	err := uc.attachContractData(ctx, job.Receipt)
	if err != nil {
		return err
	}

	chain, err := uc.db.Chain().FindOneByUUID(ctx, job.ChainUUID, []string{multitenancy.WildcardTenant}, multitenancy.WildcardOwner)
	if err != nil {
		return err
	}
	for idx, l := range job.Receipt.Logs {
		job.Receipt.Logs[idx], err = uc.decodeLogUC.Execute(ctx, chain.ChainID.String(), l)
		if err != nil {
			return err
		}
	}

	msg := job.TxResponse()
	b, err := encoding.Marshal(msg)
	if err != nil {
		errMsg := "failed to encode mined job response"
		logger.WithError(err).Error(errMsg)
		return errors.EncodingError(errMsg)
	}

	// Send message
	_, _, err = uc.kafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: uc.topicsCfg.Decoded,
		Key:   sarama.StringEncoder(job.ChainUUID),
		Value: sarama.ByteEncoder(b),
	})
	if err != nil {
		errMsg := "could not produce kafka message"
		logger.WithError(err).Error(errMsg)
		return errors.KafkaConnectionError(errMsg)
	}

	logger.WithField("msg_id", msg.Id).
		WithField("topic", uc.topicsCfg.Decoded).
		Debug("mined job notification was sent successfully")

	return nil
}

func (uc *notifyMinedJobUseCase) attachContractData(ctx context.Context, receipt *ethereum.Receipt) error {
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

	logger := uc.logger.WithContext(ctx)
	eventContract, err := uc.searchContractUC.Execute(ctx, nil, contractAddress)
	if err != nil {
		errMsg := "failed to search contract"
		logger.WithError(err).Error(errMsg)
		return errors.DependencyFailureError(errMsg)

	}

	if eventContract == nil {
		return nil
	}

	receipt.ContractName = eventContract.Name
	receipt.ContractTag = eventContract.Tag

	logger.WithField("contract_address", receipt.ContractAddress).
		WithField("contract_name", eventContract.Name).
		WithField("contract_tag", eventContract.Tag).
		Debug("source contract has been identified")
	return nil
}
