package manager

import (
	"context"
	"strings"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-sender/store"
)

const component = "nonce-manager"
const fetchNonceErr = "cannot retrieve fetch nonce from chain"

type Manager struct {
	nonce            store.NonceSender
	ethClient        ethclient.MultiClient
	recovery         store.RecoveryTracker
	maxRecovery      uint64
	chainRegistryURL string
	logger           *log.Logger
}

func NewNonceManager(ec ethclient.MultiClient, nm store.NonceSender, tracker store.RecoveryTracker, chainRegistryURL string,
	maxRecovery uint64) *Manager {
	return &Manager{
		nonce:            nm,
		ethClient:        ec,
		recovery:         tracker,
		maxRecovery:      maxRecovery,
		chainRegistryURL: chainRegistryURL,
		logger:           log.NewLogger().SetComponent(component),
	}
}

func (nc *Manager) GetNonce(ctx context.Context, job *entities.Job) (uint64, error) {
	logger := nc.logger.WithContext(ctx).WithField("job", job.UUID)

	nonceKey := job.PartitionKey()
	if nonceKey == "" {
		logger.Debug("empty nonceKey, skip nonce check")
		return 0, nil
	}

	lastSent, err := nc.nonce.GetLastSent(nonceKey)
	switch {
	case err != nil && errors.IsNotFoundError(err):
		pendingNonce, der := nc.fetchNonceFromChain(ctx, job)
		if der != nil {
			logger.WithError(der).Error(fetchNonceErr)
			return 0, der
		}

		logger.WithField("pending_nonce", pendingNonce).WithField("account", job.Transaction.From.Hex()).Debug("fetched pending nonce from node")
		return pendingNonce, nil
	case err != nil:
		errMsg := "cannot retrieve last sent nonce"
		logger.WithError(err).Error(errMsg)
		return 0, err
	default:
		return lastSent + 1, nil
	}
}

func (nc *Manager) CleanNonce(ctx context.Context, job *entities.Job, jobErr error) error {
	logger := nc.logger.WithContext(ctx).WithField("job", job.UUID)

	if job.InternalData.ParentJobUUID == job.UUID {
		logger.Debug("ignored nonce errors in children jobs")
		return nil
	}

	// TODO: update EthClient to process and standardize nonce too low errors
	if !strings.Contains(strings.ToLower(jobErr.Error()), "nonce too low") &&
		!strings.Contains(strings.ToLower(jobErr.Error()), "incorrect nonce") &&
		!strings.Contains(strings.ToLower(jobErr.Error()), "replacement transaction") {
		return nil
	}

	nonceKey := job.PartitionKey()
	logger.Warn("chain responded with invalid nonce error")
	if nc.recovery.Recovering(job.UUID) >= nc.maxRecovery {
		err := errors.InvalidNonceError("reached max nonce recovery max")
		logger.WithError(err).Error("cannot recover from nonce error")
		return err
	}

	txNonce := uint64(0)
	if job.Transaction.Nonce != nil {
		txNonce = *job.Transaction.Nonce
	}

	// Clean nonce value only if it was used to set the txNonce
	lastSentNonce, err := nc.nonce.GetLastSent(nonceKey)
	if err != nil && !errors.IsNotFoundError(err) {
		errMsg := "cannot retrieve nonce from cache for cleanup"
		logger.WithError(err).Error(errMsg)
		return err
	}

	if err == nil && txNonce == lastSentNonce+1 {
		logger.WithField("last_sent", lastSentNonce).Debug("cleaning account nonce")
		if err := nc.nonce.DeleteLastSent(nonceKey); err != nil {
			logger.WithError(err).Error("cannot clean Manager LastSent")
			return err
		}
	}

	// In case of failing because "nonce too low" we reset tx nonce
	nc.recovery.Recover(job.UUID)
	job.Transaction.Nonce = nil

	return errors.InvalidNonceWarning(jobErr.Error())
}

func (nc *Manager) IncrementNonce(ctx context.Context, job *entities.Job) error {
	logger := nc.logger.WithContext(ctx).WithField("job", job.UUID)

	nonceKey := job.PartitionKey()
	txNonce := uint64(0)
	if job.Transaction.Nonce != nil {
		txNonce = *job.Transaction.Nonce
	}

	// Set nonce value only if txNonce was using previous value
	lastSentNonce, err := nc.nonce.GetLastSent(nonceKey)
	if err != nil && !errors.IsNotFoundError(err) {
		logger.WithError(err).Error("cannot retrieve nonce from cache for increment")
		return err
	}

	if errors.IsNotFoundError(err) || txNonce == lastSentNonce+1 {
		err := nc.nonce.SetLastSent(nonceKey, txNonce)
		if err != nil {
			logger.WithError(err).Error("could not store last sent nonce")
			return err
		}
	}

	logger.WithField("last_sent", *job.Transaction.Nonce).Debug("increment account nonce value")
	nc.recovery.Recovered(job.UUID)
	return nil
}

func (nc *Manager) fetchNonceFromChain(ctx context.Context, job *entities.Job) (n uint64, err error) {
	url := client.GetProxyURL(nc.chainRegistryURL, job.ChainUUID)

	switch {
	case job.Type == entities.EEAPrivateTransaction && job.Transaction.PrivacyGroupID != "":
		n, err = nc.ethClient.PrivNonce(ctx, url, *job.Transaction.From,
			job.Transaction.PrivacyGroupID)
	case job.Type == entities.EEAPrivateTransaction && job.Transaction.PrivateFor != nil:
		n, err = nc.ethClient.PrivEEANonce(ctx, url, *job.Transaction.From,
			job.Transaction.PrivateFrom, job.Transaction.PrivateFor)
	default:
		n, err = nc.ethClient.PendingNonceAt(ctx, url, *job.Transaction.From)
	}

	return
}
