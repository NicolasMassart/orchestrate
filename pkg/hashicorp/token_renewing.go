package hashicorp

import (
	"sync"
	"time"

	"github.com/ConsenSys/orchestrate/pkg/log"
	"github.com/hashicorp/vault/api"

	"github.com/ConsenSys/orchestrate/pkg/errors"
)

// renewTokenLoop handle the token renewal of the application
type renewTokenLoop struct {
	ttl           int
	quit          chan bool
	client        *api.Client
	mut           *sync.Mutex
	retryInterval int
	maxRetries    int
	logger        *log.Logger
}

func newRenewTokenLoop(tokenExpireIn64 int64, client *api.Client, logger *log.Logger) *renewTokenLoop {
	return &renewTokenLoop{
		ttl:           int(tokenExpireIn64),
		quit:          make(chan bool, 1),
		client:        client,
		retryInterval: 2,
		maxRetries:    3,
		mut:           &sync.Mutex{},
		logger:        logger,
	}
}

// Refresh the token
func (loop *renewTokenLoop) Refresh() error {
	nRetries := 0
	for {
		newTokenSecret, err := loop.client.Auth().Token().RenewSelf(0)
		if err != nil {
			loop.logger.WithError(err).Warn("failed to refresh token")
			nRetries++
			if nRetries > loop.maxRetries {
				errMessage := "reached max number of retries to renew vault token"
				loop.logger.WithField("retries", nRetries).Error(errMessage)
				return errors.InternalError(errMessage)
			}

			time.Sleep(time.Duration(loop.retryInterval) * time.Second)
		} else {
			loop.mut.Lock()
			loop.client.SetToken(newTokenSecret.Auth.ClientToken)
			loop.mut.Unlock()
			secret, _ := loop.client.Auth().Token().LookupSelf()
			loop.logger.WithField("ttl", secret.Data["ttl"]).Info("token was refreshed successfully")
			return nil
		}
	}
}

// Run contains the token regeneration routine
func (loop *renewTokenLoop) Run() {
	go func() {
		timeToWait := time.Duration(
			int(float64(loop.ttl)*0.75), // We wait 75% of the TTL to refresh
		) * time.Second

		// Max token refresh loop of 1h
		if timeToWait > time.Hour {
			loop.logger.Info("forcing token refresh to maximum one hour")
			timeToWait = time.Hour
		}

		ticker := time.NewTicker(timeToWait)
		defer ticker.Stop()

		loop.logger.Infof("token refresh loop started (every %d seconds)", timeToWait/time.Second)
		for {
			select {
			case <-ticker.C:
				err := loop.Refresh()
				if err != nil {
					loop.quit <- true
				}

			// TODO: Be able to graceful shutdown every other services in the infra
			case <-loop.quit:
				// The token parameter is ignored
				_ = loop.client.Auth().Token().RevokeSelf("this parameter is ignored")
				// Erase the local value of the token
				loop.mut.Lock()
				loop.client.SetToken("")
				loop.mut.Unlock()
				// Wait 5 seconds for the ongoing requests to return
				time.Sleep(time.Duration(5) * time.Second)
				// Crash the tx-signer to force restart
				loop.logger.Fatal("gracefully shutting down the vault client, the token has been revoked")
			}
		}
	}()
}
