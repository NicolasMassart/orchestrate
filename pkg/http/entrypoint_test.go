package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	traefikstatic "github.com/containous/traefik/v2/pkg/config/static"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/http/handler/mock"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/http/router"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/http/router/static"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/metrics/generic"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/tls/generate"
)

type okHandler struct {
	next http.Handler
}

func (h *okHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.next.ServeHTTP(rw, req)
	rw.WriteHeader(http.StatusOK)
}

func TestEntryPoints(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	httpHandler := mock.NewMockHandler(ctrlr)
	httpsHandler := mock.NewMockHandler(ctrlr)

	// Prepare TLS configuration
	cert, err := generate.DefaultCertificate()
	require.NoError(t, err)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}

	// Create Router Configuration
	confs := map[string]*router.Router{
		"test-ep": {
			HTTP:      &okHandler{httpHandler},
			HTTPS:     &okHandler{httpsHandler},
			TLSConfig: tlsConfig,
		},
	}

	eps := NewEntryPoints(
		map[string]*traefikstatic.EntryPoint{
			"test-ep": {
				Address: "127.0.0.1:0",
				Transport: &traefikstatic.EntryPointsTransport{
					RespondingTimeouts: &traefikstatic.RespondingTimeouts{},
					LifeCycle:          &traefikstatic.LifeCycle{},
				},
			},
		},
		static.NewBuilder(confs),
		generic.NewTCP(),
	)
	_ = eps.Switch(context.Background(), nil)

	done := make(chan struct{})
	go func() {
		_ = eps.ListenAndServe(context.Background())
		close(done)
	}()

	// Wait a few millisecond for server to start
	time.Sleep(50 * time.Millisecond)

	url := eps.Addresses()["test-ep"]
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Test calling HTTP
	httpHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any())
	resp, err := client.Get(fmt.Sprintf("http://%v", url))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response should have correct status")

	// Test calling HTTPS
	httpsHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any())
	resp, err = client.Get(fmt.Sprintf("https://%v", url))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response should have correct status")

	err = eps.Shutdown(context.Background())
	assert.NoError(t, err)

	err = eps.Close()
	assert.NoError(t, err)
	<-done
}

func TestEntryPointsError(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()
	httpHandler := mock.NewMockHandler(ctrlr)

	// Create Router Configuration
	confs := map[string]*router.Router{
		"test-ep1": {HTTP: &okHandler{httpHandler}},
		"test-ep2": {HTTP: &okHandler{httpHandler}},
	}

	// We try to open 2 entry points on the same IP
	eps := NewEntryPoints(
		map[string]*traefikstatic.EntryPoint{
			"test-ep1": {
				Address: "127.0.0.1:10",
				Transport: &traefikstatic.EntryPointsTransport{
					RespondingTimeouts: &traefikstatic.RespondingTimeouts{},
					LifeCycle:          &traefikstatic.LifeCycle{},
				},
			},
			"test-ep2": {
				Address: "127.0.0.1:10",
				Transport: &traefikstatic.EntryPointsTransport{
					RespondingTimeouts: &traefikstatic.RespondingTimeouts{},
					LifeCycle:          &traefikstatic.LifeCycle{},
				},
			},
		},
		static.NewBuilder(confs),
		generic.NewTCP(),
	)
	_ = eps.Switch(context.Background(), nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	errors := eps.ListenAndServe(ctx)
	select {
	case <-errors:
	case <-time.After(time.Second):
		t.Errorf("Entrypoints should have error")
	}

	err := eps.Shutdown(ctx)
	assert.NoError(t, err)

	err = eps.Close()
	assert.NoError(t, err)
}
