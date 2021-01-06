package faucets

import (
	"context"
	"math/big"
	"reflect"

	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/chain-registry/store/models"

	log "github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/ethclient"
	types "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/types/chainregistry"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/types/entities"
	usecases "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/business/use-cases"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/business/use-cases/faucets/controls"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/chain-registry/client"
)

const getFaucetCandidateComponent = "use-cases.faucet-candidate"

type FaucetControl interface {
	Control(ctx context.Context, req *types.Request) error
	OnSelectedCandidate(ctx context.Context, faucet *entities.Faucet, beneficiary string) error
}

// RegisterContract is a use case to register a new contract
type faucetCandidate struct {
	chainRegistryClient client.ChainRegistryClient
	chainStateReader    ethclient.ChainStateReader
	searchFaucets       usecases.SearchFaucetsUseCase
	controls            []FaucetControl
}

// NewGetFaucetCandidateUseCase creates a new GetFaucetCandidateUseCase
func NewGetFaucetCandidateUseCase(
	chainRegistryClient client.ChainRegistryClient,
	searchFaucets usecases.SearchFaucetsUseCase,
	chainStateReader ethclient.ChainStateReader,
) usecases.GetFaucetCandidateUseCase {
	cooldownCtrl := controls.NewCooldownControl()
	maxBalanceCtrl := controls.NewMaxBalanceControl(chainStateReader)
	creditorCtrl := controls.NewCreditorControl(chainStateReader)

	return &faucetCandidate{
		chainRegistryClient: chainRegistryClient,
		chainStateReader:    chainStateReader,
		searchFaucets:       searchFaucets,
		controls:            []FaucetControl{creditorCtrl, cooldownCtrl, maxBalanceCtrl},
	}
}

func (uc *faucetCandidate) Execute(ctx context.Context, account string, chain *models.Chain, tenants []string) (*entities.Faucet, error) {
	logger := log.WithContext(ctx).
		WithField("chain_uuid", chain.UUID).
		WithField("account", account).
		WithField("tenants", tenants)
	logger.Debug("getting faucet candidate")

	faucets, err := uc.searchFaucets.Execute(ctx, &entities.FaucetFilters{ChainRule: chain.UUID}, tenants)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getFaucetCandidateComponent)
	}

	if len(faucets) == 0 {
		errMessage := "no faucet candidate found"
		logger.Debug(errMessage)
		return nil, errors.NotFoundError(errMessage).ExtendComponent(getFaucetCandidateComponent)
	}

	candidates := make(map[string]*entities.Faucet)
	for _, faucet := range faucets {
		candidates[faucet.UUID] = faucet
	}
	req := &types.Request{
		Beneficiary: account,
		Candidates:  candidates,
		Chain:       chain,
	}
	for _, ctrl := range uc.controls {
		err = ctrl.Control(ctx, req)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(getFaucetCandidateComponent)
		}
	}

	if len(req.Candidates) == 0 {
		errMessage := "no faucet candidate retained"
		logger.Debug(errMessage)
		return nil, errors.NotFoundError(errMessage).ExtendComponent(getFaucetCandidateComponent)
	}

	// Select a first faucet candidate for comparison
	selectedFaucet := req.Candidates[electFaucet(req.Candidates)]
	for _, ctrl := range uc.controls {
		err := ctrl.OnSelectedCandidate(ctx, selectedFaucet, req.Beneficiary)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(getFaucetCandidateComponent)
		}
	}

	logger.WithField("creditor_account", selectedFaucet.CreditorAccount).Debug("faucet candidate found successfully")
	return selectedFaucet, nil
}

// electFaucet is currently selecting the remaining faucet candidates with the highest amount
func electFaucet(faucetsCandidates map[string]*entities.Faucet) string {
	electedFaucet := reflect.ValueOf(faucetsCandidates).MapKeys()[0].String()
	amountElectedFaucetBigInt, _ := new(big.Int).SetString(faucetsCandidates[electedFaucet].Amount, 10)

	for key, candidate := range faucetsCandidates {
		amountBigInt, _ := new(big.Int).SetString(candidate.Amount, 10)

		if amountBigInt.Cmp(amountElectedFaucetBigInt) > 0 {
			electedFaucet = key
		}
	}

	return electedFaucet
}