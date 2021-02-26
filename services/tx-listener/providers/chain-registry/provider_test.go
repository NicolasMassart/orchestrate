// +build unit

package chainregistry

import (
	"context"
	"fmt"
	"github.com/ConsenSys/orchestrate/pkg/sdk/client/mock"
	"github.com/ConsenSys/orchestrate/pkg/types/api"
	"github.com/ConsenSys/orchestrate/pkg/types/testutils"
	"github.com/ConsenSys/orchestrate/pkg/utils"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ConsenSys/orchestrate/services/tx-listener/dynamic"
)

type ProviderTestSuite struct {
	suite.Suite
	provider *Provider
	client   *mock.MockOrchestrateClient
}

func (s *ProviderTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.client = mock.NewMockOrchestrateClient(ctrl)

	s.provider = &Provider{
		client: s.client,
		conf: &Config{
			RefreshInterval: time.Millisecond,
			ProxyURL:        "http://test-proxy",
		},
	}
}

func (s *ProviderTestSuite) TestRun() {
	mockChains := []*api.ChainResponse{testutils.FakeChainResponse()}

	gomock.InOrder(
		s.client.EXPECT().SearchChains(gomock.Any(), gomock.Any()).Return([]*api.ChainResponse{}, nil),
		s.client.EXPECT().SearchChains(gomock.Any(), gomock.Any()).Return(mockChains, nil).AnyTimes(),
		s.client.EXPECT().SearchChains(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes(),
	)

	cancelableCtx, cancel := context.WithCancel(context.Background())
	providerConfigUpdateCh := make(chan *dynamic.Message)
	go func() {
		runErr := s.provider.Run(cancelableCtx, providerConfigUpdateCh)
		assert.NoError(s.T(), runErr)
		close(providerConfigUpdateCh)
	}()
	config := <-providerConfigUpdateCh
	assert.Equal(s.T(), "chain-registry", config.Provider, "Should get the correct providerName")
	assert.Len(s.T(), config.Configuration.Chains, 0)

	config = <-providerConfigUpdateCh
	assert.Equal(s.T(), "chain-registry", config.Provider, "Should get the correct providerName")
	assert.Len(s.T(), config.Configuration.Chains, 1)
	assert.Equal(
		s.T(),
		utils.GetProxyURL(s.provider.conf.ProxyURL, mockChains[0].UUID),
		config.Configuration.Chains[mockChains[0].UUID].URL,
		"Chain URL should be correct",
	)

	cancel()
}

func TestProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}
