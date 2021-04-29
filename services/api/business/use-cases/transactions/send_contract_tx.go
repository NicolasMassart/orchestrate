package transactions

import (
	"context"

	"github.com/ConsenSys/orchestrate/pkg/ethereum/abi"
	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	"github.com/ConsenSys/orchestrate/pkg/utils"
	usecases "github.com/ConsenSys/orchestrate/services/api/business/use-cases"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/pkg/toolkit/app/log"
)

const sendContractTxComponent = "use-cases.send-contract-tx"

type sendContractTxUseCase struct {
	sendTxUseCase usecases.SendTxUseCase
	logger        *log.Logger
}

// NewSendContractTxUseCase creates a n¬ew SendContractTxUseCase
func NewSendContractTxUseCase(sendTxUseCase usecases.SendTxUseCase) usecases.SendContractTxUseCase {
	return &sendContractTxUseCase{
		sendTxUseCase: sendTxUseCase,
		logger:        log.NewLogger().SetComponent(sendContractTxComponent),
	}
}

// Execute validates, creates and starts a new contract transaction
func (uc *sendContractTxUseCase) Execute(ctx context.Context, txRequest *entities.TxRequest, tenantID string) (*entities.TxRequest, error) {
	ctx = log.WithFields(ctx, log.Field("idempotency-key", txRequest.IdempotencyKey))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("creating new contract transaction")

	txData, err := uc.computeTxData(txRequest.Params.MethodSignature, txRequest.Params.Args)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendContractTxComponent)
	}

	return uc.sendTxUseCase.Execute(ctx, txRequest, txData, tenantID)
}

func (uc *sendContractTxUseCase) computeTxData(method string, args []interface{}) (string, error) {
	crafter := abi.BaseCrafter{}
	sArgs, err := utils.ParseIArrayToStringArray(args)
	if err != nil {
		errMessage := "failed to parse method arguments"
		uc.logger.WithError(err).WithField("method", method).WithField("args", args).Error(errMessage)
		return "", errors.DataCorruptedError(errMessage).ExtendComponent(sendContractTxComponent)
	}

	txDataBytes, err := crafter.CraftCall(method, sArgs...)

	if err != nil {
		errMessage := "invalid method signature"
		uc.logger.WithError(err).WithField("method", method).WithField("args", args).Error(errMessage)
		return "", errors.InvalidParameterError(errMessage)
	}

	return hexutil.Encode(txDataBytes), nil
}
