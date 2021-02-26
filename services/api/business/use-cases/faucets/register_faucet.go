package faucets

import (
	"context"

	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/pkg/log"
	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	"github.com/ConsenSys/orchestrate/services/api/business/parsers"
	usecases "github.com/ConsenSys/orchestrate/services/api/business/use-cases"
	"github.com/ConsenSys/orchestrate/services/api/store"
)

const registerFaucetComponent = "use-cases.register-faucet"

// registerFaucetUseCase is a use case to register a new faucet
type registerFaucetUseCase struct {
	db             store.DB
	searchFaucetUC usecases.SearchFaucetsUseCase
	logger         *log.Logger
}

// NewRegisterFaucetUseCase creates a new RegisterFaucetUseCase
func NewRegisterFaucetUseCase(db store.DB, searchFaucetUC usecases.SearchFaucetsUseCase) usecases.RegisterFaucetUseCase {
	return &registerFaucetUseCase{
		db:             db,
		searchFaucetUC: searchFaucetUC,
		logger:         log.NewLogger().SetComponent(registerFaucetComponent),
	}
}

// Execute registers a new faucet
func (uc *registerFaucetUseCase) Execute(ctx context.Context, faucet *entities.Faucet) (*entities.Faucet, error) {
	ctx = log.WithFields(ctx, log.Field("faucet_name", faucet.Name), log.Field("chain", faucet.ChainRule))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("registering new faucet")

	faucetsRetrieved, err := uc.searchFaucetUC.Execute(ctx, &entities.FaucetFilters{
		Names: []string{faucet.Name},
	}, []string{faucet.TenantID})
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(registerFaucetComponent)
	}

	if len(faucetsRetrieved) > 0 {
		errMessage := "faucet with same name already exists"
		logger.Error(errMessage)
		return nil, errors.AlreadyExistsError(errMessage).ExtendComponent(registerFaucetComponent)
	}

	faucetModel := parsers.NewFaucetModelFromEntity(faucet)
	err = uc.db.Faucet().Insert(ctx, faucetModel)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(registerFaucetComponent)
	}

	logger.WithField("faucet_uuid", faucetModel.UUID).Info("faucet registered successfully")
	return parsers.NewFaucetFromModel(faucetModel), nil
}
