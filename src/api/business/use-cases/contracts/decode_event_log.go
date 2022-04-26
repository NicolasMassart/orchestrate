package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store"

	"github.com/consensys/orchestrate/pkg/ethereum/abi"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
)

const decodeLogsComponent = "use-cases.decode-logs"

type decodeEventLogUseCase struct {
	db                  store.DB
	getContractEventsUC usecases.GetContractEventsUseCase
	logger              *log.Logger
}

func NewDecodeEventLogUseCase(db store.DB, getContractEventsUC usecases.GetContractEventsUseCase) usecases.DecodeEventLogUseCase {
	return &decodeEventLogUseCase{
		db:                  db,
		getContractEventsUC: getContractEventsUC,
		logger:              log.NewLogger().SetComponent(decodeLogsComponent),
	}
}

func (uc *decodeEventLogUseCase) Execute(ctx context.Context, chainUUID string, eventLog *ethereum.Log) (*ethereum.Log, error) {
	chain, der := uc.db.Chain().FindOneByUUID(ctx, chainUUID, []string{multitenancy.WildcardTenant}, multitenancy.WildcardOwner)
	if der != nil {
		return nil, errors.FromError(der).ExtendComponent(decodeLogsComponent)
	}

	// This scenario is not supposed to happen
	if len(eventLog.GetTopics()) == 0 {
		errMsg := "invalid receipt (no topics in log)"
		uc.logger.Errorf(errMsg)
		return nil, errors.EncodingError(errMsg).ExtendComponent(decodeLogsComponent)
	}

	logger := uc.logger.WithContext(ctx).
		WithField("tx_hash", eventLog.GetTxHash()).
		WithField("sig_hash", eventLog.Topics[0]).
		WithField("address", eventLog.GetAddress()).
		WithField("indexed", uint32(len(eventLog.Topics)-1))

	logger.Debug("decoding receipt logs")
	sigHash := hexutil.MustDecode(eventLog.Topics[0])
	contractEvent, contractDefaultEvents, err := uc.getContractEventsUC.Execute(
		ctx,
		chain.ChainID.String(),
		ethcommon.HexToAddress(eventLog.GetAddress()),
		sigHash,
		uint32(len(eventLog.Topics)-1),
	)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}

		logger.WithError(err).Error("failed to decode receipt logs")
		return nil, errors.FromError(err).ExtendComponent(decodeLogsComponent)
	}

	if contractEvent == "" && len(contractDefaultEvents) == 0 {
		logger.Debug("could not retrieve event ABI")
		return nil, nil
	}

	var mapping map[string]string
	event := &ethAbi.Event{}

	if contractEvent != "" {
		err = json.Unmarshal([]byte(contractEvent), event)
		if err != nil {
			logger.WithError(err).Warn("could not unmarshal event")
			return nil, nil
		}
		mapping, err = abi.Decode(event, eventLog)
	} else {
		for _, potentialEvent := range contractDefaultEvents {
			// Try to unmarshal
			err = json.Unmarshal([]byte(potentialEvent), event)
			if err != nil {
				// If it fails to unmarshal, try the next potential event
				logger.WithError(err).Debug("could not unmarshal potential event")
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
		logger.WithError(err).Debug("could not unmarshal potential event")
		return nil, nil
	}

	// Set decoded data on log
	eventLog.DecodedData = mapping
	eventLog.Event = getAbi(event)

	logger.WithField("receipt_log", fmt.Sprintf("%v", mapping)).Trace("log decoded")
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
