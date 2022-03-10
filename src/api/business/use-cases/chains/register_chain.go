package chains

import (
	"context"
	"math/big"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
)

const registerChainComponent = "use-cases.register-chain"

// registerChainUseCase is a use case to register a new chain
type registerChainUseCase struct {
	db             store.DB
	searchChainsUC usecases.SearchChainsUseCase
	ethClient      ethclient.Client
	logger         *log.Logger
}

// NewRegisterChainUseCase creates a new RegisterChainUseCase
func NewRegisterChainUseCase(db store.DB, searchChainsUC usecases.SearchChainsUseCase, ec ethclient.Client) usecases.RegisterChainUseCase {
	return &registerChainUseCase{
		db:             db,
		searchChainsUC: searchChainsUC,
		ethClient:      ec,
		logger:         log.NewLogger().SetComponent(registerChainComponent),
	}
}

// Execute registers a new chain
func (uc *registerChainUseCase) Execute(ctx context.Context, chain *entities.Chain, fromLatest bool, userInfo *multitenancy.UserInfo) (*entities.Chain, error) {
	ctx = log.WithFields(ctx, log.Field("chain_name", chain.Name))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("registering new chain")

	chains, err := uc.searchChainsUC.Execute(ctx,
		&entities.ChainFilters{Names: []string{chain.Name}, TenantID: userInfo.TenantID},
		userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(registerChainComponent)
	}

	if len(chains) > 0 {
		errMessage := "a chain with the same name already exists"
		logger.Error(errMessage)
		return nil, errors.AlreadyExistsError(errMessage).ExtendComponent(registerChainComponent)
	}

	chainID, err := uc.getChainID(ctx, chain.URLs)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(registerChainComponent)
	}
	chain.ChainID = chainID

	if fromLatest {
		chainTip, der := uc.getChainTip(ctx, chain.URLs)
		if der != nil {
			return nil, errors.FromError(der).ExtendComponent(registerChainComponent)
		}

		chain.ListenerStartingBlock = chainTip
		chain.ListenerCurrentBlock = chainTip
	}

	chain.TenantID = userInfo.TenantID
	chain.OwnerID = userInfo.Username
	err = uc.db.Chain().Insert(ctx, chain)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(registerChainComponent)
	}

	logger.WithField("chain_uuid", chain.UUID).Info("chain registered successfully")
	return chain, nil
}

func (uc *registerChainUseCase) getChainID(ctx context.Context, uris []string) (*big.Int, error) {
	var prevChainID *big.Int
	for i, uri := range uris {
		chainID, err := uc.ethClient.Network(ctx, uri)
		if err != nil {
			errMessage := "failed to fetch chain id"
			uc.logger.WithContext(ctx).WithField("url", uri).WithError(err).Error(errMessage)
			return nil, errors.InvalidParameterError(errMessage)
		}

		if i > 0 && chainID.String() != prevChainID.String() {
			errMessage := "URLs in the list point to different networks"
			uc.logger.WithContext(ctx).
				WithField("url", uri).
				WithField("previous_chain_id", prevChainID).
				WithField("chain_id", chainID.String()).
				Error(errMessage)
			return nil, errors.InvalidParameterError(errMessage)
		}

		prevChainID = chainID
	}

	return prevChainID, nil
}

func (uc *registerChainUseCase) getChainTip(ctx context.Context, uris []string) (uint64, error) {
	for _, uri := range uris {
		header, err := uc.ethClient.HeaderByNumber(ctx, uri, nil)
		if err != nil {
			errMessage := "failed to fetch chain tip"
			uc.logger.WithContext(ctx).WithField("url", uri).WithError(err).Warn(errMessage)
			continue
		}

		return header.Number.Uint64(), nil
	}

	errMessage := "failed to fetch chain tip for all urls"
	uc.logger.WithContext(ctx).WithField("uris", uris).Error(errMessage)
	return 0, errors.InvalidParameterError(errMessage)
}
