// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient/utils"
	pkgutils "github.com/consensys/orchestrate/tests/pkg/utils"
	"github.com/consensys/quorum-key-manager/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type chainTestSuite struct {
	suite.Suite
	env    *Environment
	ctx    context.Context
	cancel context.CancelFunc
}

func TestChains(t *testing.T) {
	s := new(chainTestSuite)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	defer s.cancel()

	var err error
	s.env, err = NewEnvironment(s.ctx, s.cancel)
	require.NoError(t, err)

	sig := common.NewSignalListener(func(signal os.Signal) {
		s.env.Logger.Error("interrupt signal was caught")
		s.cancel()
		t.FailNow()
	})

	defer sig.Close()

	suite.Run(t, s)
}

func (s *chainTestSuite) TestChains_SuccessfulUserStories() {
	s.T().Run("as a user I want to create chain and perform RPC calls through", func(t *testing.T) {
		regChainReq := testdata.FakeRegisterChainRequest()
		regChainReq.PrivateTxManagerURL = ""
		regChainReq.URLs = s.env.TestData.Nodes.Geth[0].URLs
		res, err := s.env.Client.RegisterChain(s.ctx, regChainReq)
		require.NoError(t, err)

		defer func() {
			err = s.env.Client.DeleteChain(s.ctx, res.UUID)
			require.NoError(t, err)
		}()

		assert.Equal(t, s.env.UserInfo.TenantID, res.TenantID)
		assert.Equal(t, regChainReq.Name, res.Name)
		assert.Equal(t, regChainReq.URLs, res.URLs)
		assert.Equal(t, regChainReq.Labels, res.Labels)
		assert.Equal(t, regChainReq.Listener.Depth, res.ListenerDepth)
		assert.Equal(t, regChainReq.Listener.BlockTimeDuration, res.ListenerBlockTimeDuration)

		chainProxyURL := s.env.Client.ChainProxyURL(res.UUID)
		err = pkgutils.WaitForProxy(s.ctx, chainProxyURL, s.env.EthClient)
		require.NoError(t, err)

		cID, err := s.env.EthClient.Network(s.ctx, chainProxyURL)
		require.NoError(t, err)
		assert.NotEmpty(t, cID.String())
	})

	s.T().Run("as a user I want to create chain, update and search it", func(t *testing.T) {
		regChainReq := testdata.FakeRegisterChainRequest()
		regChainReq.PrivateTxManagerURL = ""
		regChainReq.URLs = s.env.TestData.Nodes.Geth[0].URLs
		res, err := s.env.Client.RegisterChain(s.ctx, regChainReq)
		require.NoError(t, err)

		defer func() {
			err = s.env.Client.DeleteChain(s.ctx, res.UUID)
			require.NoError(t, err)
		}()

		updateChainReq := testdata.FakeUpdateChainRequest()
		res2, err := s.env.Client.UpdateChain(s.ctx, res.UUID, updateChainReq)
		require.NoError(t, err)
		assert.Equal(t, updateChainReq.Name, res2.Name)
		assert.Equal(t, updateChainReq.Labels, res2.Labels)
		assert.Equal(t, updateChainReq.Listener.Depth, res2.ListenerDepth)
		assert.Equal(t, updateChainReq.Listener.BlockTimeDuration, res2.ListenerBlockTimeDuration)
		require.NoError(t, err)

		res3, err := s.env.Client.SearchChains(s.ctx, &entities.ChainFilters{
			Names: []string{updateChainReq.Name},
		})
		require.NoError(t, err)
		assert.Equal(t, res2, res3[0])
	})

	s.T().Run("as a user I want to store in cache block data responses when rpc nodes are the same", func(t *testing.T) {
		regChainReq := testdata.FakeRegisterChainRequest()
		regChainReq.PrivateTxManagerURL = ""
		regChainReq.URLs = s.env.TestData.Nodes.Geth[0].URLs
		res, err := s.env.Client.RegisterChain(s.ctx, regChainReq)
		require.NoError(t, err)
		defer func() {
			err = s.env.Client.DeleteChain(s.ctx, res.UUID)
			require.NoError(t, err)
		}()

		regChainReq2 := testdata.FakeRegisterChainRequest()
		regChainReq2.PrivateTxManagerURL = ""
		regChainReq2.URLs = s.env.TestData.Nodes.Geth[0].URLs
		res2, err := s.env.Client.RegisterChain(s.ctx, regChainReq2)
		require.NoError(t, err)
		defer func() {
			err = s.env.Client.DeleteChain(s.ctx, res2.UUID)
			require.NoError(t, err)
		}()

		params, _ := json.Marshal([]interface{}{"0x1", false})
		rpcMsg := &utils.JSONRpcMessage{
			Method:  "eth_getBlockByNumber",
			Version: "2.0",
			Params:  params,
			ID:      strconv.AppendUint(nil, uint64(1), 10),
		}
		body, _ := json.Marshal(rpcMsg)

		chainProxyURL := s.env.Client.ChainProxyURL(res.UUID)
		err = pkgutils.WaitForProxy(s.ctx, chainProxyURL, s.env.EthClient)
		require.NoError(t, err)
		rpcRes, err := s.env.HTTPClient.Post(chainProxyURL, "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		var respMsg utils.JSONRpcMessage
		err = json.NewDecoder(rpcRes.Body).Decode(&respMsg)
		require.NoError(t, err)

		chainProxyURL2 := s.env.Client.ChainProxyURL(res2.UUID)
		err = pkgutils.WaitForProxy(s.ctx, chainProxyURL2, s.env.EthClient)
		require.NoError(t, err)
		rpcRes2, err := s.env.HTTPClient.Post(chainProxyURL2, "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		var respMsg2 utils.JSONRpcMessage
		err = json.NewDecoder(rpcRes2.Body).Decode(&respMsg2)
		require.NoError(t, err)

		assert.Equal(t, respMsg, respMsg2)
		assert.NotEmpty(t, rpcRes2.Header.Get("X-Cache-Control"))

		// Wait for cache to disable
		time.Sleep(time.Second * 3)

		rpcRes3, err := s.env.HTTPClient.Post(chainProxyURL2, "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		var respMsg3 utils.JSONRpcMessage
		err = json.NewDecoder(rpcRes3.Body).Decode(&respMsg3)
		require.NoError(t, err)
		assert.Equal(t, respMsg, respMsg3)
		assert.Empty(t, rpcRes3.Header.Get("X-Cache-Control"))

	})
}

func (s *chainTestSuite) TestChains_FailureScenarios() {
	s.T().Run("when an user creates two chains with same name if should fail with expected error", func(t *testing.T) {
		regChainReq := testdata.FakeRegisterChainRequest()
		regChainReq.PrivateTxManagerURL = ""
		regChainReq.URLs = s.env.TestData.Nodes.Geth[0].URLs
		res, err := s.env.Client.RegisterChain(s.ctx, regChainReq)
		require.NoError(t, err)
		defer func() {
			err = s.env.Client.DeleteChain(s.ctx, res.UUID)
			require.NoError(t, err)
		}()

		_, err = s.env.Client.RegisterChain(s.ctx, regChainReq)
		assert.Error(t, err)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())
	})

	s.T().Run("when an user creates a chains with invalid argument expected error", func(t *testing.T) {
		regChainReq := testdata.FakeRegisterChainRequest()
		regChainReq.URLs = []string{"aSDASDASD"}
		_, err := s.env.Client.RegisterChain(s.ctx, regChainReq)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())

		regChainReq = testdata.FakeRegisterChainRequest()
		regChainReq.URLs = s.env.TestData.Nodes.Geth[0].URLs
		regChainReq.Listener.BlockTimeDuration = "-12as"
		_, err = s.env.Client.RegisterChain(s.ctx, regChainReq)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})
}
