//go:build integration
// +build integration

package integrationtests

import (
	"net/http"
	"testing"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type eventStreamsTestSuite struct {
	suite.Suite
	client client.OrchestrateClient
	env    *IntegrationEnvironment
	chain  *types.ChainResponse
}

func (s *eventStreamsTestSuite) SetupSuite() {
	ctx := s.env.ctx

	chainReq := testdata.FakeRegisterChainRequest()
	chainReq.URLs = []string{s.env.blockchainNodeURL}
	chainReq.PrivateTxManagerURL = ""

	var err error
	s.chain, err = s.client.RegisterChain(ctx, chainReq)
	require.NoError(s.T(), err)
}

func (s *eventStreamsTestSuite) TestCreate() {
	ctx := s.env.ctx

	s.T().Run("should create event stream successfully: Webhook", func(t *testing.T) {
		req := testdata.FakeCreateWebhookEventStreamRequest()
		req.Chain = s.chain.Name

		resp, err := s.client.CreateWebhookEventStream(ctx, req)
		require.NoError(t, err)

		assert.Equal(t, s.chain.UUID, resp.ChainUUID)
		assert.Equal(t, req.Labels, resp.Labels)
		assert.Equal(t, string(entities.EventStreamChannelWebhook), resp.Channel)
		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, string(entities.EventStreamStatusLive), resp.Status)
		assert.NotEmpty(t, resp.UUID)
		assert.NotEmpty(t, resp.CreatedAt)
		assert.NotEmpty(t, resp.UpdatedAt)

		err = s.client.DeleteEventStream(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should create event stream successfully: Kafka", func(t *testing.T) {
		req := testdata.FakeCreateKafkaEventStreamRequest()
		req.Chain = s.chain.Name

		resp, err := s.client.CreateKafkaEventStream(ctx, req)
		require.NoError(t, err)

		assert.Equal(t, s.chain.UUID, resp.ChainUUID)
		assert.Equal(t, req.Labels, resp.Labels)
		assert.Equal(t, string(entities.EventStreamChannelKafka), resp.Channel)
		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, string(entities.EventStreamStatusLive), resp.Status)
		assert.NotEmpty(t, resp.UUID)
		assert.NotEmpty(t, resp.CreatedAt)
		assert.NotEmpty(t, resp.UpdatedAt)

		err = s.client.DeleteEventStream(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should fail to register event stream with BadRequest if chain does not exists", func(t *testing.T) {
		req := testdata.FakeCreateWebhookEventStreamRequest()
		req.Chain = "invalidChain"

		_, err := s.client.CreateWebhookEventStream(ctx, req)
		require.Error(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail to register overlapped event stream per tenant", func(t *testing.T) {
		req := testdata.FakeCreateWebhookEventStreamRequest()
		req.Chain = ""

		resp, err := s.client.CreateWebhookEventStream(ctx, req)
		require.NoError(t, err)

		_, err = s.client.CreateWebhookEventStream(ctx, req)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())

		err = s.client.DeleteEventStream(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should fail to register event stream with same name and tenant", func(t *testing.T) {
		req := testdata.FakeCreateWebhookEventStreamRequest()
		req.Chain = s.chain.Name

		resp, err := s.client.CreateWebhookEventStream(ctx, req)
		require.NoError(t, err)

		_, err = s.client.CreateWebhookEventStream(ctx, req)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())

		err = s.client.DeleteEventStream(ctx, resp.UUID)
		assert.NoError(t, err)
	})

	s.T().Run("should fail to register event stream with same chain and tenant", func(t *testing.T) {
		req := testdata.FakeCreateWebhookEventStreamRequest()
		req.Name = utils.RandString(5)
		req.Chain = s.chain.Name

		resp, err := s.client.CreateWebhookEventStream(ctx, req)
		require.NoError(t, err)

		_, err = s.client.CreateWebhookEventStream(ctx, req)
		req.Name = utils.RandString(5)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())

		err = s.client.DeleteEventStream(ctx, resp.UUID)
		assert.NoError(t, err)
	})
}

func (s *eventStreamsTestSuite) TestSearch() {
	chainReq := testdata.FakeRegisterChainRequest()
	chainReq.URLs = []string{s.env.blockchainNodeURL}
	chainReq.PrivateTxManagerURL = ""
	chain, err := s.client.RegisterChain(s.env.ctx, chainReq)
	require.NoError(s.T(), err)

	ctx := s.env.ctx
	req := testdata.FakeCreateWebhookEventStreamRequest()
	req.Chain = chain.Name

	es, err := s.client.CreateWebhookEventStream(ctx, req)
	require.NoError(s.T(), err)

	defer func() {
		err = s.client.DeleteEventStream(ctx, es.UUID)
		assert.NoError(s.T(), err)
	}()

	s.T().Run("should search event stream by name successfully", func(t *testing.T) {
		resp, err := s.client.SearchEventStreams(ctx, &entities.EventStreamFilters{
			Names: []string{es.Name},
		})
		require.NoError(t, err)

		assert.Len(t, resp, 1)
		assert.Equal(t, es.UUID, resp[0].UUID)
	})

	s.T().Run("should search event stream by chain_uuid successfully", func(t *testing.T) {
		resp, err := s.client.SearchEventStreams(ctx, &entities.EventStreamFilters{
			ChainUUID: es.ChainUUID,
		})
		require.NoError(t, err)

		assert.Len(t, resp, 1)
		assert.Equal(t, es.UUID, resp[0].UUID)
	})
}

func (s *eventStreamsTestSuite) TestGetOne() {
	ctx := s.env.ctx
	req := testdata.FakeCreateKafkaEventStreamRequest()
	req.Chain = s.chain.Name

	es, err := s.client.CreateKafkaEventStream(ctx, req)
	require.NoError(s.T(), err)

	defer func() {
		err = s.client.DeleteEventStream(ctx, es.UUID)
		require.NoError(s.T(), err)
	}()

	s.T().Run("should get event stream successfully", func(t *testing.T) {
		resp, err := s.client.GetEventStream(ctx, es.UUID)
		require.NoError(t, err)
		assert.Equal(t, es.UUID, resp.UUID)
	})
}

func (s *eventStreamsTestSuite) TestUpdate() {
	ctx := s.env.ctx

	s.T().Run("should update event stream successfully: Webhook", func(t *testing.T) {
		req := testdata.FakeCreateWebhookEventStreamRequest()
		req.Chain = s.chain.Name
		esWebhook, err := s.client.CreateWebhookEventStream(ctx, req)
		require.NoError(s.T(), err)

		resp, err := s.client.UpdateWebhookEventStream(ctx, esWebhook.UUID, testdata.FakeUpdateWebhookEventStreamRequest())
		require.NoError(t, err)

		assert.Equal(t, string(entities.EventStreamStatusSuspend), resp.Status)
		assert.NotEmpty(t, resp.UUID)
		assert.True(t, resp.UpdatedAt.After(resp.CreatedAt))

		err = s.client.DeleteEventStream(ctx, esWebhook.UUID)
		assert.NoError(s.T(), err)
	})

	s.T().Run("should update event stream successfully: Kafka", func(t *testing.T) {
		req2 := testdata.FakeCreateKafkaEventStreamRequest()
		req2.Chain = s.chain.Name
		esKafka, err := s.client.CreateKafkaEventStream(ctx, req2)
		require.NoError(s.T(), err)

		resp, err := s.client.UpdateKafkaEventStream(ctx, esKafka.UUID, testdata.FakeUpdateKafkaEventStreamRequest())
		require.NoError(t, err)

		assert.Equal(t, string(entities.EventStreamStatusSuspend), resp.Status)
		assert.NotEmpty(t, resp.UUID)
		assert.True(t, resp.UpdatedAt.After(resp.CreatedAt))

		err = s.client.DeleteEventStream(ctx, esKafka.UUID)
		assert.NoError(s.T(), err)
	})
}
