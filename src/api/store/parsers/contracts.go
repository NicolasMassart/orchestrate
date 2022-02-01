package parsers

import (
	"encoding/json"

	ethabi "github.com/consensys/orchestrate/pkg/ethereum/abi"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TODO: Remove this function as parsing the events from the ABI should not be done on Orchestrate as we do not have control on how the events are represented in the ABI

// ParseEvents returns a map of events given an ABI
func ParseEvents(data string) (map[string]string, error) {
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

func NewArtifactModel(artifact *entities.Artifact) *models.ArtifactModel {
	return &models.ArtifactModel{
		ABI:              artifact.ABI,
		Bytecode:         artifact.Bytecode.String(),
		DeployedBytecode: artifact.DeployedBytecode.String(),
		Codehash:         artifact.CodeHash.String(),
	}
}

func NewArtifactEntity(artifact *models.ArtifactModel) *entities.Artifact {
	res := &entities.Artifact{
		ABI:              artifact.ABI,
		Bytecode:         hexutil.MustDecode(artifact.Bytecode),
		DeployedBytecode: hexutil.MustDecode(artifact.DeployedBytecode),
	}

	if artifact.Codehash != "" {
		res.CodeHash = hexutil.MustDecode(artifact.Codehash)
	}

	return res
}

func NewEventModel(event *entities.ContractEvent) *models.EventModel {
	return &models.EventModel{
		ABI:               event.ABI,
		SigHash:           event.SigHash.String(),
		IndexedInputCount: event.IndexedInputCount,
		Codehash:          event.CodeHash.String(),
	}
}

func NewEventModelArr(events []*entities.ContractEvent) []*models.EventModel {
	res := []*models.EventModel{}
	for _, e := range events {
		res = append(res, NewEventModel(e))
	}
	return res
}

func NewEventEntity(event *models.EventModel) *entities.ContractEvent {
	res := &entities.ContractEvent{
		ABI:               event.ABI,
		IndexedInputCount: event.IndexedInputCount,
	}

	if event.SigHash != "" {
		res.SigHash = hexutil.MustDecode(event.SigHash)
	}

	if event.Codehash != "" {
		res.CodeHash = hexutil.MustDecode(event.Codehash)
	}

	return res
}

func NewEventEntityArr(events []*models.EventModel) []*entities.ContractEvent {
	res := []*entities.ContractEvent{}
	for _, e := range events {
		res = append(res, NewEventEntity(e))
	}
	return res
}
