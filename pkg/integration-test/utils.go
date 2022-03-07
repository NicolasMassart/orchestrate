package integrationtest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/traefik/traefik/v2/pkg/log"
)

func WaitForServiceLive(ctx context.Context, url, name string, timeout time.Duration) {
	logger := log.FromContext(ctx)
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	errMsg := fmt.Sprintf("cannot reach %s service on %s", name, url)
	for {
		req, _ := http.NewRequest("GET", url, nil)
		req = req.WithContext(rctx)

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			if resp != nil && resp.StatusCode == 200 {
				logger.Infof("service %s is live", name)
				return
			}

			logger.WithField("status", resp.StatusCode).Warnf(errMsg)
		}

		if rctx.Err() != nil {
			logger.WithError(rctx.Err()).Warnf(errMsg)
			return
		}

		logger.Debugf("waiting for 1 s for service %s to start...", name)
		time.Sleep(time.Second)
	}
}
