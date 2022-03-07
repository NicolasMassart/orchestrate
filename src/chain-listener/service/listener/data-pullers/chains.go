package datapullers

import (
	"context"
	"sync"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	usecases "github.com/consensys/orchestrate/src/chain-listener/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/chain-listener/service/listener"
	"github.com/consensys/orchestrate/src/chain-listener/service/listener/events"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	uuid2 "github.com/google/uuid"
)

const chainsComponent = "chain-listener.listener.chains"

type chainsService struct {
	client           orchestrateclient.OrchestrateClient
	ethClient        ethclient.Client
	interval         time.Duration
	logger           *log.Logger
	subscriptions    map[string]chan *events.Chain
	lastCheckAt      time.Time
	activeChains     map[string]*entities.Chain
	addChainEvent    usecases.AddChainUseCase
	updateChainEvent usecases.UpdateChainUseCase
	deleteChainEvent usecases.DeleteChainUseCase
	cancelCtx        context.CancelFunc
	cerr             chan error
}

func ChainListenerService(client orchestrateclient.OrchestrateClient,
	ethClient ethclient.Client,
	addChainEvent usecases.AddChainUseCase,
	updateChainEvent usecases.UpdateChainUseCase,
	deleteChainEvent usecases.DeleteChainUseCase,
	interval time.Duration,
	logger *log.Logger,
) listener.ChainListener {
	return &chainsService{
		client:           client,
		interval:         interval,
		ethClient:        ethClient,
		addChainEvent:    addChainEvent,
		updateChainEvent: updateChainEvent,
		deleteChainEvent: deleteChainEvent,
		subscriptions:    make(map[string]chan *events.Chain),
		activeChains:     make(map[string]*entities.Chain),
		logger:           logger.SetComponent(chainsComponent),
		cerr:             make(chan error, 1),
	}
}

func (s *chainsService) Run(ctx context.Context) error {
	s.logger.WithField("refresh", s.interval.String()).Info("chains listener started")
	ctx, s.cancelCtx = context.WithCancel(ctx)

	err := s.runIt(ctx)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			<-ticker.C
			if err = s.runIt(ctx); err != nil {
				if ctx.Err() == nil { // Context is not done
					s.cerr <- err
				}
				return
			}
		}
	}()

	select {
	case err = <-s.cerr:
		s.logger.WithField("err", err).Info("chains listener exited with errors")
		return err
	case <-ctx.Done():
		s.logger.WithField("reason", ctx.Err().Error()).Info("chains listener gracefully stopping...")
		return nil
	}
}

func (s *chainsService) runIt(ctx context.Context) error {
	chainEvents, err := s.retrieveChains(ctx)
	if err != nil {
		s.logger.WithField("err", err).Info("chain listener exit with errors")
		return errors.FromError(err).ExtendComponent(chainsComponent)
	}

	s.notifySubscribers(ctx, chainEvents)

	for _, event := range chainEvents {
		switch event.Type {
		case events.NewChain:
			if err := s.addChainEvent.Execute(ctx, event.Chain); err != nil {
				return err
			}
			s.activeChains[event.Chain.UUID] = event.Chain
		case events.UpdatedChain:
			if err := s.updateChainEvent.Execute(ctx, event.Chain); err != nil {
				return err
			}
			s.activeChains[event.Chain.UUID] = event.Chain
		case events.DeletedChain:
			if err := s.deleteChainEvent.Execute(ctx, event.Chain.UUID); err != nil {
				return err
			}
			delete(s.activeChains, event.Chain.UUID)
		}
	}

	return nil
}

func (s *chainsService) Subscribe(sub chan *events.Chain) string {
	subID := uuid2.New().String()
	s.subscriptions[subID] = sub
	return subID
}

func (s *chainsService) Unsubscribe(subID string) error {
	if sub, ok := s.subscriptions[subID]; ok {
		close(sub)
		delete(s.subscriptions, subID)
		return nil
	}

	return errors.InvalidParameterError("there is not active subscription %s", subID)
}

func (s *chainsService) Close() error {
	defer close(s.cerr)
	s.cancelCtx()
	for subID := range s.subscriptions {
		_ = s.Unsubscribe(subID)
	}

	s.logger.Info("chains listener closed")
	return nil
}

func (s *chainsService) notifySubscribers(_ context.Context, chainEvents []*events.Chain) {
	if len(chainEvents) == 0 {
		return
	}

	wg := &sync.WaitGroup{}
	for _, sub := range s.subscriptions {
		wg.Add(1)
		go func(in chan *events.Chain) {
			for _, e := range chainEvents {
				in <- e
			}
			wg.Done()
		}(sub)
	}

	wg.Wait()
	s.logger.Debug("subscribers has been notified")
}

func (s *chainsService) retrieveChains(ctx context.Context) ([]*events.Chain, error) {
	chainResponses, err := s.client.SearchChains(ctx, &entities.ChainFilters{})
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch pending jobs")
		return nil, err
	}

	activeChainUUIDs := map[string]bool{}
	chainEvents := []*events.Chain{}
	nCreateChainEvents := 0
	nUpdatedChainEvents := 0
	nDeletedChainEvents := 0
	for _, chainResponse := range chainResponses {
		activeChainUUIDs[chainResponse.UUID] = true
		if _, ok := s.activeChains[chainResponse.UUID]; ok {
			if chainResponse.UpdatedAt.After(s.lastCheckAt) {
				nUpdatedChainEvents++
				chainEvents = append(chainEvents, &events.Chain{
					Chain: formatters.ChainResponseToEntity(chainResponse),
					Type:  events.UpdatedChain,
				})
			}
		} else {
			nCreateChainEvents++
			chainEvents = append(chainEvents, &events.Chain{
				Chain: formatters.ChainResponseToEntity(chainResponse),
				Type:  events.NewChain,
			})
		}
	}

	for chainUUID, chain := range s.activeChains {
		if _, ok := activeChainUUIDs[chainUUID]; !ok {
			nDeletedChainEvents++
			chainEvents = append(chainEvents, &events.Chain{
				Chain: chain,
				Type:  events.DeletedChain,
			})
		}
	}

	s.lastCheckAt = time.Now()
	s.logger.
		WithField("created", nCreateChainEvents).
		WithField("updated", nUpdatedChainEvents).
		WithField("deleted", nDeletedChainEvents).
		Debug("retrieved chains")
	return chainEvents, nil
}
