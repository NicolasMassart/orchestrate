package contracts

import (
	"bytes"
	"context"
	"encoding/json"

	ethabi "github.com/consensys/orchestrate/pkg/ethereum/abi"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

const registerContractComponent = "use-cases.register-contract"

type registerContractUseCase struct {
	db     store.DB
	logger *log.Logger
}

func NewRegisterContractUseCase(agent store.DB) usecases.RegisterContractUseCase {
	return &registerContractUseCase{
		db:     agent,
		logger: log.NewLogger().SetComponent(registerContractComponent),
	}
}

func (uc *registerContractUseCase) Execute(ctx context.Context, contract *entities.Contract) error {
	ctx = log.WithFields(ctx, log.Field("contract_id", contract))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("registering contract starting...")

	retrievedContract, err := uc.db.Contract().FindOneByNameAndTag(ctx, contract.Name, contract.Tag)
	if err != nil && !errors.IsNotFoundError(err) {
		return errors.FromError(err).ExtendComponent(registerContractComponent)
	}

	contract.RawABI, err = getABICompacted(contract.RawABI)
	if err != nil {
		errMessage := "failed to parse contract ABI"
		logger.WithError(err).Error(errMessage)
		return errors.InvalidParameterError(errMessage).ExtendComponent(registerContractComponent)
	}
	contract.CodeHash = crypto.Keccak256(contract.DeployedBytecode)

	events, err := getEvents(contract)
	if err != nil {
		errMessage := "failed to parse contract events from ABI"
		logger.WithError(err).Error(errMessage)
		return errors.InvalidParameterError(errMessage).ExtendComponent(registerContractComponent)
	}

	if retrievedContract == nil {
		err = uc.db.Contract().Register(ctx, contract)
	} else {
		err = uc.db.Contract().Update(ctx, contract)
	}
	if err != nil {
		return errors.FromError(err).ExtendComponent(registerContractComponent)
	}

	if len(events) > 0 {
		err = uc.db.ContractEvent().RegisterMultiple(ctx, events)
		if err != nil {
			return errors.FromError(err).ExtendComponent(registerContractComponent)
		}
	}

	logger.Info("contract registered successfully")
	return nil
}

func getEvents(contract *entities.Contract) ([]entities.ContractEvent, error) {
	eventJSONs, err := parseEvents(contract.RawABI)
	if err != nil {
		return nil, err
	}
	var events []entities.ContractEvent
	// nolint
	for _, e := range contract.ABI.Events {
		indexedCount := getIndexedCount(&e)
		if contract.DeployedBytecode != nil {
			events = append(events, entities.ContractEvent{
				CodeHash:          contract.CodeHash,
				SigHash:           e.ID.Bytes(),
				IndexedInputCount: indexedCount,
				ABI:               eventJSONs[e.Sig],
			})
		}
	}

	return events, nil
}

func getABICompacted(rawABI string) (string, error) {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(rawABI)); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// returns the count of indexed inputs in the event
func getIndexedCount(event *abi.Event) (indexedInputCount uint) {
	for i := range event.Inputs {
		if event.Inputs[i].Indexed {
			indexedInputCount++
		}
	}

	return indexedInputCount
}

// TODO: Remove this function as parsing the events from the ABI should not be done on Orchestrate as we do not have control on how the events are represented in the ABI

func parseEvents(data string) (map[string]string, error) {
	var parsedFields []entities.RawABI

	err := json.Unmarshal([]byte(data), &parsedFields)
	if err != nil {
		return nil, err
	}

	// Retrieve raw JSONs
	normalizedJSON, err := json.Marshal(parsedFields)
	if err != nil {
		return nil, err
	}

	var rawFields []json.RawMessage
	err = json.Unmarshal(normalizedJSON, &rawFields)
	if err != nil {
		return nil, err
	}

	events := make(map[string]string)
	for i := 0; i < len(rawFields) && i < len(parsedFields); i++ {
		fieldJSON, err := rawFields[i].MarshalJSON()
		if err != nil {
			return nil, err
		}

		if parsedFields[i].Type == "event" {
			e := &ethabi.Event{}
			err := json.Unmarshal(fieldJSON, e)
			if err != nil {
				return nil, err
			}

			events[e.Name+e.Sig()] = string(fieldJSON)
		}
	}

	return events, nil
}
