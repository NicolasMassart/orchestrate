package postgres

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/postgres"
)

type PGContractEvent struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.ContractEventAgent = &PGContractEvent{}

func NewPGContractEvent(client postgres.Client) *PGContractEvent {
	return &PGContractEvent{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.contract-event"),
	}
}

func (agent *PGContractEvent) RegisterMultiple(ctx context.Context, events []entities.ContractEvent) error {
	var eventModels []models.Event
	for _, e := range events {
		eventModels = append(eventModels, *models.NewEvent(&e))
	}

	err := agent.client.ModelContext(ctx, &eventModels).
		OnConflict("DO NOTHING").
		Insert()
	if err != nil {
		errMessage := "failed to insert multiple contract event"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	return nil
}

func (agent *PGContractEvent) FindOneByAccountAndSigHash(ctx context.Context, chainID, address, sighash string, indexedInputCount uint32) (*entities.ContractEvent, error) {
	event := &models.Event{}
	err := agent.client.ModelContext(ctx, event).
		Column("event.abi").
		Join("JOIN codehashes AS c ON c.codehash = event.codehash").
		Where("c.chain_id = ?", chainID).
		Where("c.address = ?", address).
		Where("event.sig_hash = ?", sighash).
		Where("event.indexed_input_count = ?", indexedInputCount).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}

		errMessage := "failed to find contract event by address and signature hash"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return event.ToEntity(), nil
}

func (agent *PGContractEvent) FindDefaultBySigHash(ctx context.Context, sighash string, indexedInputCount uint32) ([]*entities.ContractEvent, error) {
	var defaultEvents []models.Event
	err := agent.client.ModelContext(ctx, &defaultEvents).
		ColumnExpr("DISTINCT abi").
		Where("sig_hash = ?", sighash).
		Where("indexed_input_count = ?", indexedInputCount).
		Order("abi DESC").
		Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMessage := "failed to find default contract events by signature hash"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return models.NewEvents(defaultEvents), nil
}
