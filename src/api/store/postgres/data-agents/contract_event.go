package dataagents

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
)

const contractEventDAComponent = "data-agents.contract-event"

type PGContractEvent struct {
	db     pg.DB
	logger *log.Logger
}

func NewPGContractEvent(db pg.DB) store.ContractEventAgent {
	return &PGContractEvent{
		db:     db,
		logger: log.NewLogger().SetComponent(contractDAComponent),
	}
}

func (agent *PGContractEvent) RegisterMultiple(ctx context.Context, events []*entities.ContractEvent) error {
	eventModels := parsers.NewEventModelArr(events)
	query := agent.db.ModelContext(ctx, &eventModels).OnConflict("DO NOTHING")
	err := pg.InsertQuery(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert multiple contract event")
		return errors.FromError(err).ExtendComponent(contractEventDAComponent)
	}

	return nil
}

func (agent *PGContractEvent) FindOneByAccountAndSigHash(ctx context.Context, chainID, address, sighash string, indexedInputCount uint32) (*entities.ContractEvent, error) {
	event := &models.EventModel{}
	query := agent.db.ModelContext(ctx, event).
		Column("event_model.abi").
		Join("JOIN codehashes AS c ON c.codehash = event_model.codehash").
		Where("c.chain_id = ?", chainID).
		Where("c.address = ?", address).
		Where("event_model.sig_hash = ?", sighash).
		Where("event_model.indexed_input_count = ?", indexedInputCount)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to find event")
		}
		return nil, errors.FromError(err).ExtendComponent(contractEventDAComponent)
	}

	return parsers.NewEventEntity(event), nil
}

func (agent *PGContractEvent) FindDefaultBySigHash(ctx context.Context, sighash string, indexedInputCount uint32) ([]*entities.ContractEvent, error) {
	var defaultEvents []*models.EventModel
	query := agent.db.ModelContext(ctx, &defaultEvents).
		ColumnExpr("DISTINCT abi").
		Where("sig_hash = ?", sighash).
		Where("indexed_input_count = ?", indexedInputCount).
		Order("abi DESC")

	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to find default event")
		}
		return nil, errors.FromError(err).ExtendComponent(contractEventDAComponent)
	}

	return parsers.NewEventEntityArr(defaultEvents), nil
}
