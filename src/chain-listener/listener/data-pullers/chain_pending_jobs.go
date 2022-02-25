package datapullers

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	"github.com/consensys/orchestrate/src/chain-listener/listener"
	"github.com/consensys/orchestrate/src/chain-listener/listener/events"
	usecases "github.com/consensys/orchestrate/src/chain-listener/use-cases"
	"github.com/consensys/orchestrate/src/entities"
)

const chainPendingTxsComponent = "chain-listener.data-puller.chain-pending-jobs"
const chainPendingTxsSessionComponent = "chain-listener.data-puller.chain-pending-jobs.session"

type chainPendingJobsListener struct {
	client                 orchestrateclient.OrchestrateClient
	logger                 *log.Logger
	chainListener          listener.ChainListener
	onChainPendingJobEvent usecases.PendingJobEventHandler
	cancelCtx              context.CancelFunc
	sessions               map[string]*chainPendingJobsSession
	cerr                   chan error
}

func ChainPendingJobsListener(client orchestrateclient.OrchestrateClient,
	chainListener listener.ChainListener,
	onChainPendingJobEvent usecases.PendingJobEventHandler,
	logger *log.Logger,
) listener.ChainPendingJobsListener {
	return &chainPendingJobsListener{
		client:                 client,
		chainListener:          chainListener,
		onChainPendingJobEvent: onChainPendingJobEvent,
		logger:                 logger.SetComponent(chainPendingTxsComponent),
		sessions:               make(map[string]*chainPendingJobsSession),
		cerr:                   make(chan error, 1),
	}
}

func (l *chainPendingJobsListener) Run(ctx context.Context) error {
	l.logger.Info("chain pending job session builder started")
	ctx, l.cancelCtx = context.WithCancel(ctx)

	go func() {
		chainEvents := make(chan *events.Chain, 1) // @TODO Optimize chan size
		defer close(chainEvents)
		subscriptionID := l.chainListener.Subscribe(chainEvents)
		for event := range chainEvents {
			switch event.Type {
			case events.NewChain:
				err := l.addChainSession(ctx, event.Chain)
				if err != nil && !errors.IsAlreadyExistsError(err) {
					l.cerr <- err
					break
				}
			case events.UpdatedChain:
				err := l.updateChainSession(ctx, event.Chain)
				if err != nil {
					l.cerr <- err
					break
				}
			case events.DeletedChain:
				err := l.removeChainSession(ctx, event.Chain)
				if err != nil {
					l.cerr <- err
					break
				}
			}
		}

		_ = l.chainListener.Unsubscribe(subscriptionID)
	}()

	select {
	case err := <-l.cerr:
		l.logger.WithField("err", err).Info("chain pending job session builder exited with errors")
		return err
	case <-ctx.Done():
		l.logger.WithField("reason", ctx.Err().Error()).Info("chain pending job session builder gracefully stopping...")
		return nil
	}
}

func (l *chainPendingJobsListener) Close() error {
	l.cancelCtx()
	gerr := []error{}
	for _, sess := range l.sessions {
		if err := sess.Stop(); err != nil {
			gerr = append(gerr, err)
		}
	}

	// @TODO Concatenate errors
	if len(gerr) > 1 {
		return gerr[0]
	}

	return nil
}

func (l *chainPendingJobsListener) addChainSession(ctx context.Context, chain *entities.Chain) error {
	if _, ok := l.sessions[chain.UUID]; ok {
		errMsg := "chain pending job session already exist"
		l.logger.WithField("chain", chain.UUID).Error(errMsg)
		return errors.AlreadyExistsError(errMsg)
	}

	sess := newChainPendingJobSession(l.client, l.onChainPendingJobEvent, chain, l.logger)
	l.sessions[chain.UUID] = sess
	go func(s *chainPendingJobsSession) {
		err := s.Start(ctx)
		if err != nil {
			l.cerr <- err
		}
	}(sess)
	return nil
}

func (l *chainPendingJobsListener) removeChainSession(_ context.Context, chain *entities.Chain) error {
	l.logger.Debug("removing chain pending job session...")
	if sess, ok := l.sessions[chain.UUID]; ok {
		return sess.Stop()
	}

	errMsg := "session not exist"
	l.logger.WithField("chain", chain.UUID).Error(errMsg)
	return errors.NotFoundError(errMsg)
}

func (l *chainPendingJobsListener) updateChainSession(_ context.Context, chain *entities.Chain) error {
	l.logger.Debug("updating chain pending job session...")
	sessID := chain.UUID
	if sess, ok := l.sessions[sessID]; ok {
		return sess.Update(chain)
	}

	errMsg := "chain pending job session does not exist"
	l.logger.WithField("chain", chain.UUID).Error(errMsg)
	return errors.NotFoundError(errMsg)
}

type chainPendingJobsSession struct {
	client                 orchestrateclient.OrchestrateClient
	onChainPendingJobEvent usecases.PendingJobEventHandler
	lastCallAt             time.Time
	logger                 *log.Logger
	chain                  *entities.Chain
	cancelCtx              context.CancelFunc
	ticker                 *time.Ticker
	cerr                   chan error
}

func newChainPendingJobSession(client orchestrateclient.OrchestrateClient,
	onChainPendingJobEvent usecases.PendingJobEventHandler,
	chain *entities.Chain,
	logger *log.Logger,
) *chainPendingJobsSession {
	return &chainPendingJobsSession{
		client:                 client,
		chain:                  chain,
		onChainPendingJobEvent: onChainPendingJobEvent,
		logger:                 logger.WithField("chain", chain.UUID).SetComponent(chainPendingTxsSessionComponent),
	}
}

func (s *chainPendingJobsSession) Start(ctx context.Context) error {
	s.logger.Info("chain pending job listening started")
	ctx, s.cancelCtx = context.WithCancel(ctx)

	err := s.runIt(ctx)
	if err != nil {
		return err
	}

	go func() {
		s.ticker = time.NewTicker(s.chain.ListenerBackOffDuration)
		defer s.ticker.Stop()
		for {
			<-s.ticker.C
			err := s.runIt(ctx)
			if err != nil {
				s.cerr <- err
				return
			}
		}
	}()

	select {
	case err := <-s.cerr:
		s.logger.WithField("err", err).Info("chain pending job listener exited with errors")
		return err
	case <-ctx.Done():
		s.logger.WithField("reason", ctx.Err().Error()).Info("chain pending job listener gracefully stopping...")
		return nil
	}
}

func (s *chainPendingJobsSession) Update(chain *entities.Chain) error {
	s.chain = chain
	s.ticker.Reset(s.chain.ListenerBackOffDuration)
	s.logger.Debug("chain pending job listener updated")
	return nil
}

func (s *chainPendingJobsSession) Stop() error {
	s.cancelCtx()
	s.logger.Info("chain pending job listener closed")
	return nil
}

func (s *chainPendingJobsSession) runIt(ctx context.Context) error {
	pendingJobs, err := s.retrievePendingTxs(ctx)
	if err != nil {
		return errors.FromError(err).ExtendComponent(chainPendingTxsComponent)
	}

	// @TODO Evaluate to run in parallel
	for _, job := range pendingJobs {
		if err := s.onChainPendingJobEvent.Execute(ctx, job); err != nil {
			return err
		}
	}

	return nil
}

func (s *chainPendingJobsSession) retrievePendingTxs(ctx context.Context) ([]*entities.Job, error) {
	// We get all the pending jobs updated_after the last tick
	filters := &entities.JobFilters{
		Status:    entities.StatusPending,
		ChainUUID: s.chain.UUID,
	}

	if !s.lastCallAt.IsZero() {
		filters.UpdatedAfter = s.lastCallAt
	}

	jobResponses, err := s.client.SearchJob(ctx, filters)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch pending jobs")
		return nil, err
	}
	s.lastCallAt = time.Now()

	pendingJobEvents := make([]*entities.Job, len(jobResponses))
	for idx, jobResponse := range jobResponses {
		pendingJobEvents[idx] = formatters.JobResponseToEntity(jobResponse)
	}

	s.logger.WithField("total", len(pendingJobEvents)).
		WithField("updated_after", filters.UpdatedAfter.Format("2006-01-02 15:04:05")).
		Debug("retrieved pending jobs")
	return pendingJobEvents, nil
}
