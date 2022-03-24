package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/consensys/orchestrate/pkg/ethereum/abi"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
)

const decodeLogsComponent = "use-cases.decode-logs"

type decodeEventLogUseCase struct {
	getContractEventsUC usecases.GetContractEventsUseCase
	logger              *log.Logger
}

func NewDecodeEventLogUseCase(getContractEventsUC usecases.GetContractEventsUseCase) usecases.DecodeEventLogUseCase {
	return &decodeEventLogUseCase{
		getContractEventsUC: getContractEventsUC,
		logger:              log.NewLogger().SetComponent(decodeLogsComponent),
	}
}

func (uc *decodeEventLogUseCase) Execute(ctx context.Context, chainID string, eventLog *ethereum.Log) (*ethereum.Log, error) {
	uc.logger.Debug("decoding receipt logs...")

	// This scenario is not supposed to happen
	if len(eventLog.GetTopics()) == 0 {
		errMsg := "invalid receipt (no topics in log)"
		uc.logger.Errorf(errMsg)
		return nil, errors.EncodingError(errMsg)
	}

	logger := uc.logger.WithContext(ctx).WithField("sig_hash", utils.ShortString(eventLog.Topics[0], 5)).
		WithField("address", eventLog.GetAddress()).WithField("indexed", uint32(len(eventLog.Topics)-1))

	logger.Debug("decoding receipt logs")
	sigHash := hexutil.MustDecode(eventLog.Topics[0])
	contractEvent, contractDefaultEvents, err := uc.getContractEventsUC.Execute(
		ctx,
		chainID,
		ethcommon.HexToAddress(eventLog.GetAddress()),
		sigHash,
		uint32(len(eventLog.Topics)-1),
	)

	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}

		logger.WithError(err).Error("failed to decode receipt logs")
		return nil, err
	}

	if contractEvent == "" && len(contractDefaultEvents) == 0 {
		logger.WithError(err).Debug("could not retrieve event ABI")
		return nil, nil
	}

	var mapping map[string]string
	event := &ethAbi.Event{}

	if contractEvent != "" {
		err = json.Unmarshal([]byte(contractEvent), event)
		if err != nil {
			logger.WithError(err).
				Warnf("could not unmarshal event ABI provided by the Contract Registry, txHash: %s sigHash: %s, ", eventLog.GetTxHash(), eventLog.GetTopics()[0])
			return nil, nil
		}
		mapping, err = abi.Decode(event, eventLog)
	} else {
		for _, potentialEvent := range contractDefaultEvents {
			// Try to unmarshal
			err = json.Unmarshal([]byte(potentialEvent), event)
			if err != nil {
				// If it fails to unmarshal, try the next potential event
				logger.WithError(err).Tracef("could not unmarshal potential event ABI, txHash: %s sigHash: %s, ", eventLog.GetTxHash(), eventLog.GetTopics()[0])
				continue
			}

			// Try to decode
			mapping, err = abi.Decode(event, eventLog)
			if err == nil {
				// As the decoding is successful, stop looping
				break
			}
		}
	}

	if err != nil {
		// As all potentialEvents fail to unmarshal, go to the next log
		logger.WithError(err).Tracef("could not unmarshal potential event ABI, txHash: %s sigHash: %s, ", eventLog.GetTxHash(), eventLog.GetTopics()[0])
		return nil, nil
	}

	// Set decoded data on log
	eventLog.DecodedData = mapping
	eventLog.Event = getAbi(event)

	logger.WithField("receipt_log", fmt.Sprintf("%v", mapping)).
		Debug("log decoded")

	return eventLog, nil
}

// GetAbi creates a string ABI (format EventName(argType1, argType2)) from an event
func getAbi(e *ethAbi.Event) string {
	inputs := make([]string, len(e.Inputs))
	for i := range e.Inputs {
		inputs[i] = fmt.Sprintf("%v", e.Inputs[i].Type)
	}
	return fmt.Sprintf("%v(%v)", e.Name, strings.Join(inputs, ","))
}
