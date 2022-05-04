// +build integration

package integrationtests

import (
	"net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type chainsTestSuite struct {
	suite.Suite
	client sdk.OrchestrateClient
	env    *IntegrationEnvironment
}

func (s *chainsTestSuite) TestRegister() {
	ctx := s.env.ctx

	s.T().Run("should register chain successfully from latest", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.URLs = []string{s.env.blockchainNodeURL}

		resp, err := s.client.RegisterChain(ctx, req)
		require.NoError(t, err)

		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, req.URLs, resp.URLs)
		assert.Equal(t, multitenancy.DefaultTenant, resp.TenantID)
		assert.Equal(t, req.Listener.Depth, resp.ListenerDepth)
		assert.Equal(t, req.Labels, resp.Labels)
		assert.NotEmpty(t, resp.UUID)
		assert.NotEmpty(t, resp.CreatedAt)
		assert.NotEmpty(t, resp.UpdatedAt)
		assert.Equal(t, resp.CreatedAt, resp.UpdatedAt)

		err = s.client.DeleteChain(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should register chain successfully from latest if fromBlock is empty", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.URLs = []string{s.env.blockchainNodeURL}

		resp, err := s.client.RegisterChain(ctx, req)
		require.NoError(t, err)

		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, req.URLs, resp.URLs)
		assert.Equal(t, multitenancy.DefaultTenant, resp.TenantID)
		assert.Equal(t, req.Listener.Depth, resp.ListenerDepth)
		assert.NotEmpty(t, resp.UUID)
		assert.NotEmpty(t, resp.CreatedAt)
		assert.NotEmpty(t, resp.UpdatedAt)
		assert.Equal(t, resp.CreatedAt, resp.UpdatedAt)

		err = s.client.DeleteChain(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should register chain successfully from 0", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.URLs = []string{s.env.blockchainNodeURL}

		resp, err := s.client.RegisterChain(ctx, req)
		require.NoError(t, err)

		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, req.URLs, resp.URLs)
		assert.Equal(t, multitenancy.DefaultTenant, resp.TenantID)
		assert.Equal(t, req.Listener.Depth, resp.ListenerDepth)
		assert.NotEmpty(t, resp.UUID)
		assert.NotEmpty(t, resp.CreatedAt)
		assert.NotEmpty(t, resp.UpdatedAt)
		assert.Equal(t, resp.CreatedAt, resp.UpdatedAt)

		err = s.client.DeleteChain(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should fail with 400 if payload is invalid", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.URLs = nil

		_, err := s.client.RegisterChain(ctx, req)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail with 400 if invalid backoff duration", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.Listener.BlockTimeDuration = "invalidDuration"

		_, err := s.client.RegisterChain(ctx, req)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail with 400 if invalid urls", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.URLs = []string{"invalidURL"}

		_, err := s.client.RegisterChain(ctx, req)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail with 400 if invalid private tx manager url", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.PrivateTxManagerURL = "invalidURL"

		_, err := s.client.RegisterChain(ctx, req)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail with 422 if URL is not reachable", func(t *testing.T) {
		_, err := s.client.RegisterChain(ctx, testdata.FakeRegisterChainRequest())
		assert.Equal(t, http.StatusUnprocessableEntity, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail with 409 if chain with same name and tenant already exists", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.URLs = []string{s.env.blockchainNodeURL}

		resp, err := s.client.RegisterChain(ctx, req)
		require.NoError(t, err)

		_, err = s.client.RegisterChain(ctx, req)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())

		err = s.client.DeleteChain(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should fail with 500 to register chain if postgres is down", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()

		err := s.env.client.Stop(ctx, postgresContainerID)
		assert.NoError(t, err)

		_, err = s.client.RegisterChain(ctx, req)
		assert.Error(t, err)

		err = s.env.client.StartServiceAndWait(ctx, postgresContainerID, 10*time.Second)
		assert.NoError(t, err)
	})
}

func (s *chainsTestSuite) TestSearch() {
	ctx := s.env.ctx
	req := testdata.FakeRegisterChainRequest()
	req.URLs = []string{s.env.blockchainNodeURL}
	chain, err := s.client.RegisterChain(ctx, req)
	require.NoError(s.T(), err)

	s.T().Run("should search chain by name successfully", func(t *testing.T) {
		resp, err := s.client.SearchChains(ctx, &entities.ChainFilters{
			Names: []string{chain.Name},
		})
		require.NoError(t, err)

		assert.Len(t, resp, 1)
		assert.Equal(t, chain.UUID, resp[0].UUID)
	})

	s.T().Run("should return empty array if nothing is found", func(t *testing.T) {
		resp, err := s.client.SearchChains(ctx, &entities.ChainFilters{
			Names: []string{"inexistentName"},
		})
		require.NoError(t, err)
		assert.Empty(t, resp)
	})

	err = s.client.DeleteChain(ctx, chain.UUID)
	assert.NoError(s.T(), err)
}

func (s *chainsTestSuite) TestGetOne() {
	ctx := s.env.ctx

	s.T().Run("should get chain successfully", func(t *testing.T) {
		req := testdata.FakeRegisterChainRequest()
		req.URLs = []string{s.env.blockchainNodeURL}
		chain, err := s.client.RegisterChain(ctx, req)
		require.NoError(s.T(), err)

		resp, err := s.client.GetChain(ctx, chain.UUID)
		require.NoError(t, err)
		assert.Equal(t, chain.UUID, resp.UUID)

		err = s.client.DeleteChain(ctx, chain.UUID)
		require.NoError(s.T(), err)
	})
}

func (s *chainsTestSuite) TestUpdate() {
	ctx := s.env.ctx
	req := testdata.FakeRegisterChainRequest()
	req.URLs = []string{s.env.blockchainNodeURL}
	chain, err := s.client.RegisterChain(ctx, req)
	require.NoError(s.T(), err)

	s.T().Run("should update chain name successfully", func(t *testing.T) {
		req := testdata.FakeUpdateChainRequest()
		req.Name = "newName"

		resp, err := s.client.UpdateChain(ctx, chain.UUID, req)
		require.NoError(t, err)

		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, req.Labels, resp.Labels)
		assert.NotEqual(t, resp.CreatedAt, resp.UpdatedAt)
	})

	err = s.client.DeleteChain(ctx, chain.UUID)
	assert.NoError(s.T(), err)
}
