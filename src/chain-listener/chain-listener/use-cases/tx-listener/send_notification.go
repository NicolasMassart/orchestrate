package txlistener

import (
	"context"

	"github.com/Shopify/sarama"
	encoding "github.com/consensys/orchestrate/pkg/encoding/sarama"
	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

const sendNotificationComponent = "chain-listener.use-case.tx-listener.send-notification"

type sendNotificationUseCase struct {
	client   orchestrateclient.OrchestrateClient
	producer sarama.SyncProducer
	topic    string
	logger   *log.Logger
}

func SendNotificationUseCase(client orchestrateclient.OrchestrateClient,
	producer sarama.SyncProducer,
	topic string,
	logger *log.Logger,
) usecases.SendNotification {
	return &sendNotificationUseCase{
		client:   client,
		producer: producer,
		topic:    topic,
		logger:   logger.SetComponent(sendNotificationComponent),
	}
}

func (uc *sendNotificationUseCase) Execute(ctx context.Context, job *entities.Job) error {
	logger := uc.logger.WithField("chain", job.ChainUUID).WithField("job", job.UUID).WithField("tx_hash", job.Transaction.Hash)
	logger.Debug("sending job notification")

	err := uc.attachContractData(ctx, job.Receipt)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{}
	err = encoding.Marshal(job.TxResponse(), msg)
	if err != nil {
		return err
	}

	msg.Topic = uc.topic
	msg.Key = sarama.StringEncoder(job.ChainUUID)
	_, _, err = uc.producer.SendMessage(msg)
	if err != nil {
		logger.WithError(err).Errorf("failed to produce message")
		return err
	}

	logger.Info("notification has been sent")
	return nil
}

func (uc *sendNotificationUseCase) attachContractData(ctx context.Context, receipt *ethereum.Receipt) error {
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
	eventContract, err := uc.client.SearchContract(ctx, &types.SearchContractRequest{
		Address: contractAddress,
	})

	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil
		}
		logger.WithError(err).Error("failed to search contract")
		return err

	}

	receipt.ContractName = eventContract.Name
	receipt.ContractTag = eventContract.Tag

	logger.WithField("contract_address", receipt.ContractAddress).
		WithField("contract_name", eventContract.Name).
		WithField("contract_tag", eventContract.Tag).
		Debug("source contract has been identified")
	return nil
}
