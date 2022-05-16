// +build integration

package integrationtests

import (
	"fmt"
	http2 "net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type txListenerJobTestSuite struct {
	suite.Suite
	env *IntegrationEnvironment
}

const accOneRawTxOne = "0xf86e808443a5b73882520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820713a0c23154893b531af779316139fd21a59643ec446209bda55afb23ba9cc1ec5720a00355d30b02fe453f3e9595a4b7fbf7a290560e3e1d4f5869167ed5870e8d69a7"
const accOneTxHashOne = "0x4c280d7f7b9ae00b44b472156a898e19ee96ec07b89d770dc124734306b431ea"

const accOneRawTxTwo = "0xf86e018442a4599182520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820713a08193ae278c6f7bb0568105873ba899266110d9d8f7aca1bd5300554bd6a18a6da00d96cf498d9f2649bfdd1daecb8fe29b6a647a4cf1efde0df21d697f336a1683"
const accOneTxHashTwo = "0xb0ee44a2d19c554790ff89de85230284e2956678e4ce4967a3f237b8fa807fb2"

const accTwoRawTxOne = "0xf86e80845015d85582520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820713a0f4ed10c5f072dfb100d0167368d09dc5868007c39fb9d56becb91983e53640e2a00cb014710c24bbd58d5ad0ef672031704a3b5fb42e4f941525c6f9b5d23ae708"
const accTwoTxHashOne = "0x1f1ed4d866459192c7ffdfc09205451d0fee0bdf087ff5938b6c662bba524bd4"

const accTwoRawTxTwo = "0xf86e01844d86768b82520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820714a0f8549cb1e99b3b417dbf58b45060f02da2625e8f18d462ae478943f2f561802fa06b81c6051ae9bd1fbeed7e2ec2fca1d0287dbb674a01e8cb5d59bbdeb6924b88"
const accTwoTxHashTwo = "0xcd2ba08ae7d9f79f2b72fc0b35975a3c6303de9667cfce0d5df2e906b31f6944"

var extendedWaitingTime = time.Millisecond * 500

func (s *txListenerJobTestSuite) TestPendingJob() {
	respWaitingTime := s.env.chain.ListenerBlockTimeDuration + extendedWaitingTime

	s.T().Run("should get pending job and wait for mined transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accOneTxHashOne)).(*ethcommon.Hash)

		gock.New(apiURL).
			Get("/chains/" + ganacheChainUUID).
			Reply(http2.StatusOK).JSON(formatters.FormatChainResponse(s.env.chain))

		err := s.sendJobMessage(job)
		require.NoError(s.T(), err)

		hash, err := s.env.ethClient.SendRawTransaction(s.env.ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accOneRawTxOne))
		require.NoError(t, err)
		assert.Equal(t, hash.String(), job.Transaction.Hash.String())
		
		updateMsg, err := s.env.messengerConsumerTracker.WaitForUpdateJobMessage(s.env.ctx, job.UUID, entities.StatusMined, respWaitingTime)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), accOneTxHashOne, updateMsg.Receipt.TxHash)
	})

	s.T().Run("should get pending job and fetch available receipt right away successfully", func(t *testing.T) {
		defer gock.OffAll()
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accOneTxHashTwo)).(*ethcommon.Hash)

		gock.New(apiURL).
			Get("/chains/" + ganacheChainUUID).
			Reply(http2.StatusOK).JSON(formatters.FormatChainResponse(s.env.chain))

		hash, err := s.env.ethClient.SendRawTransaction(s.env.ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accOneRawTxTwo))
		require.NoError(t, err)
		assert.Equal(t, hash.String(), job.Transaction.Hash.String())

		err = s.sendJobMessage(job)
		require.NoError(s.T(), err)
		
		updateMsg, err := s.env.messengerConsumerTracker.WaitForUpdateJobMessage(s.env.ctx, job.UUID, entities.StatusMined, respWaitingTime)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), accOneTxHashTwo, updateMsg.Receipt.TxHash)
	})
}

func (s *txListenerJobTestSuite) TestRetryJob() {
	s.T().Run("should retry pending job and wait for mined successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accTwoTxHashOne)).(*ethcommon.Hash)
		job.InternalData.RetryInterval = s.env.chain.ListenerBlockTimeDuration * 2

		gock.New(apiURL).
			Get("/chains/" + ganacheChainUUID).
			Reply(http2.StatusOK).JSON(formatters.FormatChainResponse(s.env.chain))

		gock.New(apiURL).
			Get("/jobs").
			AddMatcher(searchTxMatcher(job.UUID)).
			Reply(http2.StatusOK).JSON([]*api.JobResponse{
			formatters.FormatJobResponse(job),
		})

		isResendCalled := make(chan bool, 1)
		gock.New(apiURL).
			Put(fmt.Sprintf("/jobs/%s/resend", job.UUID)).
			AddMatcher(waitTimeoutMatcher(isResendCalled, job.InternalData.RetryInterval+extendedWaitingTime)).
			Reply(http2.StatusAccepted)

		err := s.sendJobMessage(job)
		require.NoError(t, err)

		go func() {
			<-time.Tick(s.env.chain.ListenerBlockTimeDuration + extendedWaitingTime)
			hash, err := s.env.ethClient.SendRawTransaction(s.env.ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accTwoRawTxOne))
			require.NoError(t, err)
			assert.Equal(t, hash.String(), job.Transaction.Hash.String())
		}()

		assert.True(t, <-isResendCalled, "missing resend job call")

		// Update to MINED
		gock.New(apiURL).
			Patch("/jobs/" + job.UUID).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		isResendCalledTwice := make(chan bool, 1)
		gock.New(apiURL).
			Put(fmt.Sprintf("/jobs/%s/resend", job.UUID)).
			AddMatcher(waitTimeoutMatcher(isResendCalledTwice, job.InternalData.RetryInterval)).
			Reply(http2.StatusAccepted)
		assert.False(t, <-isResendCalledTwice, "unexpected resend job call")
	})

	s.T().Run("should send last retry of pending job successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accTwoTxHashTwo)).(*ethcommon.Hash)
		job.InternalData.RetryInterval = s.env.chain.ListenerBlockTimeDuration
		for idx := api.SentryMaxRetries; idx > 1; idx-- {
			job.Logs = append(job.Logs, &entities.Log{
				Status: entities.StatusResending,
			})
		}

		gock.New(apiURL).
			Get("/chains/" + ganacheChainUUID).
			Reply(http2.StatusOK).JSON(formatters.FormatChainResponse(s.env.chain))

		gock.New(apiURL).
			Get("/jobs").
			AddMatcher(searchTxMatcher(job.UUID)).
			Reply(http2.StatusOK).JSON([]api.JobResponse{})

		isResendCalled := make(chan bool, 1)
		gock.New(apiURL).
			Put(fmt.Sprintf("/jobs/%s/resend", job.UUID)).
			AddMatcher(waitTimeoutMatcher(isResendCalled, job.InternalData.RetryInterval+extendedWaitingTime)).
			Reply(http2.StatusAccepted)

		err := s.sendJobMessage(job)

		require.NoError(t, err)
		assert.True(t, <-isResendCalled, "missing resend job call")
		
		_, err = s.env.messengerConsumerTracker.WaitForUpdateJobMessage(s.env.ctx, job.UUID, entities.StatusMined, job.InternalData.RetryInterval+extendedWaitingTime)
		require.NoError(s.T(), err)
	})

	s.T().Run("should send new child job of pending job successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accTwoTxHashTwo)).(*ethcommon.Hash)
		job.Transaction.Raw = hexutil.MustDecode(accTwoRawTxTwo)
		job.InternalData.RetryInterval = s.env.chain.ListenerBlockTimeDuration
		job.InternalData.GasPriceIncrement = 0.01
		job.InternalData.GasPriceLimit = 0.1
		childrenJob := testdata.FakeJob()
		childrenJob.InternalData.ParentJobUUID = job.UUID

		gock.New(apiURL).
			Get("/chains/" + ganacheChainUUID).
			Reply(http2.StatusOK).JSON(formatters.FormatChainResponse(s.env.chain))

		gock.New(apiURL).
			Get("/jobs").
			AddMatcher(searchTxMatcher(job.UUID)).
			Reply(http2.StatusOK).JSON([]api.JobResponse{})

		gock.New(apiURL).
			Post("/jobs").
			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(childrenJob))

		isStartJobCalled := make(chan bool, 1)
		gock.New(apiURL).
			Put("/jobs/" + childrenJob.UUID + "/start").
			AddMatcher(waitTimeoutMatcher(isStartJobCalled, job.InternalData.RetryInterval+extendedWaitingTime)).
			Reply(http2.StatusAccepted)

		err := s.sendJobMessage(job)

		require.NoError(t, err)
		assert.True(t, <-isStartJobCalled, "missing update job call")
	})
}

func (s *txListenerJobTestSuite) sendJobMessage(job *entities.Job) error {
	s.env.logger.WithField("job", job.UUID).WithField("topic", s.env.cfg.ConsumerTopic).
		Info("sending message")
	err := s.env.messengerClient.PendingJobMessage(s.env.ctx, job, multitenancy.NewInternalAdminUser())
	if err != nil {
		return err
	}

	return nil
}

func searchTxMatcher(jobUUID string) gock.MatchFunc {
	return func(rw *http2.Request, _ *gock.Request) (bool, error) {
		qParentJobUUID := rw.URL.Query().Get("parent_job_uuid")
		if qParentJobUUID == jobUUID {
			return true, nil
		}
		return false, nil
	}
}

func waitTimeoutMatcher(c chan bool, duration time.Duration) gock.MatchFunc {
	go func() {
		<-time.After(duration)
		c <- false
	}()

	return func(rw *http2.Request, _ *gock.Request) (bool, error) {
		c <- true
		return true, nil
	}
}
