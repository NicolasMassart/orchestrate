// +build integration

package integrationtests
// 
// import (
// 	"fmt"
// 	http2 "net/http"
// 	"testing"
// 	"time"
// 
// 	"github.com/consensys/orchestrate/pkg/utils"
// 	"github.com/consensys/orchestrate/src/api/service/formatters"
// 	api "github.com/consensys/orchestrate/src/api/service/types"
// 	"github.com/consensys/orchestrate/src/entities"
// 	"github.com/consensys/orchestrate/src/entities/testdata"
// 	ethcommon "github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/common/hexutil"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// 	"gopkg.in/h2non/gock.v1"
// )
// 
// type txSentryTestSuite struct {
// 	suite.Suite
// 	env *IntegrationEnvironment
// }
// 
// // From: 0x93f7274c9059e601be4512F656B57b830e019E41
// const accTwoRawTxOne = "0xf86e80845015d85582520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820713a0f4ed10c5f072dfb100d0167368d09dc5868007c39fb9d56becb91983e53640e2a00cb014710c24bbd58d5ad0ef672031704a3b5fb42e4f941525c6f9b5d23ae708"
// const accTwoTxHashOne = "0x1f1ed4d866459192c7ffdfc09205451d0fee0bdf087ff5938b6c662bba524bd4"
// 
// const accTwoRawTxTwo = "0xf86e01844d86768b82520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820714a0f8549cb1e99b3b417dbf58b45060f02da2625e8f18d462ae478943f2f561802fa06b81c6051ae9bd1fbeed7e2ec2fca1d0287dbb674a01e8cb5d59bbdeb6924b88"
// const accTwoTxHashTwo = "0xcd2ba08ae7d9f79f2b72fc0b35975a3c6303de9667cfce0d5df2e906b31f6944"
// 
// var waitForEnvelopeTimeOut = time.Second * 5
// 
// func (s *txSentryTestSuite) TestRetryTxs() {
// 	ctx := s.env.ctx
// 	chain := testdata.FakeChain()
// 	chain.UUID = ganacheChainUUID
// 
// 	s.T().Run("should retry pending job and wait for mined successfully", func(t *testing.T) {
// 		job := testdata.FakeJob()
// 		job.ChainUUID = s.env.chain.UUID
// 		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accTwoTxHashOne)).(*ethcommon.Hash)
// 		job.InternalData.RetryInterval = time.Second * 2
// 
// 		gock.New(apiURL).
// 			Get("/jobs").
// 			AddMatcher(searchTxMatcher(job.UUID)).
// 			Reply(http2.StatusOK).JSON([]*api.JobResponse{
// 			formatters.FormatJobResponse(job),
// 		})
// 
// 		isResendCalled := make(chan bool, 1)
// 		gock.New(apiURL).
// 			Put(fmt.Sprintf("/jobs/%s/resend", job.UUID)).
// 			AddMatcher(waitTimeoutMatcher(isResendCalled, job.InternalData.RetryInterval+time.Second)).
// 			Reply(http2.StatusAccepted)
// 
// 		err := s.env.ucs.PendingJobUseCase().Execute(ctx, job)
// 
// 		require.NoError(t, err)
// 		assert.True(t, <-isResendCalled, "missing resend job call")
// 
// 		isResendCalledTwice := make(chan bool, 1)
// 		gock.New(apiURL).
// 			Put(fmt.Sprintf("/jobs/%s/resend", job.UUID)).
// 			AddMatcher(waitTimeoutMatcher(isResendCalledTwice, job.InternalData.RetryInterval)).
// 			Reply(http2.StatusAccepted)
// 
// 		// Update to MINED
// 		gock.New(apiURL).
// 			Patch("/jobs/" + job.UUID).
// 			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))
// 
// 		hash, err := s.env.ethClient.SendRawTransaction(ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accTwoRawTxOne))
// 		require.NoError(t, err)
// 		assert.Equal(t, hash.String(), job.Transaction.Hash.String())
// 		assert.False(t, <-isResendCalledTwice, "unexpected resend job call")
// 	})
// 
// 	s.T().Run("should send last retry of pending job successfully", func(t *testing.T) {
// 		job := testdata.FakeJob()
// 		job.ChainUUID = s.env.chain.UUID
// 		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accTwoTxHashTwo)).(*ethcommon.Hash)
// 		job.InternalData.RetryInterval = time.Second
// 		for idx := api.SentryMaxRetries; idx > 1; idx-- {
// 			job.Logs = append(job.Logs, &entities.Log{
// 				Status: entities.StatusResending,
// 			})
// 		}
// 
// 		gock.New(apiURL).
// 			Get("/jobs").
// 			AddMatcher(searchTxMatcher(job.UUID)).
// 			Reply(http2.StatusOK).JSON([]api.JobResponse{})
// 
// 		isResendCalled := make(chan bool, 1)
// 		gock.New(apiURL).
// 			Put(fmt.Sprintf("/jobs/%s/resend", job.UUID)).
// 			AddMatcher(waitTimeoutMatcher(isResendCalled, job.InternalData.RetryInterval+time.Second)).
// 			Reply(http2.StatusAccepted)
// 
// 		isUpdatedCalled := make(chan bool, 1)
// 		gock.New(apiURL).
// 			Patch("/jobs/" + job.UUID).
// 			AddMatcher(waitTimeoutMatcher(isUpdatedCalled, job.InternalData.RetryInterval+time.Second)).
// 			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))
// 
// 		err := s.env.ucs.PendingJobUseCase().Execute(ctx, job)
// 
// 		require.NoError(t, err)
// 		assert.True(t, <-isResendCalled, "missing resend job call")
// 		assert.True(t, <-isUpdatedCalled, "missing update job call")
// 	})
// 
// 	s.T().Run("should send new child job of pending job successfully", func(t *testing.T) {
// 		job := testdata.FakeJob()
// 		job.ChainUUID = s.env.chain.UUID
// 		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accTwoTxHashTwo)).(*ethcommon.Hash)
// 		job.Transaction.Raw = hexutil.MustDecode(accTwoRawTxTwo)
// 		job.InternalData.RetryInterval = time.Second
// 		job.InternalData.GasPriceIncrement = 0.01
// 		job.InternalData.GasPriceLimit = 0.1
// 		childrenJob := testdata.FakeJob()
// 		childrenJob.InternalData.ParentJobUUID = job.UUID
// 
// 		gock.New(apiURL).
// 			Get("/jobs").
// 			AddMatcher(searchTxMatcher(job.UUID)).
// 			Reply(http2.StatusOK).JSON([]api.JobResponse{})
// 
// 		gock.New(apiURL).
// 			Post("/jobs").
// 			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(childrenJob))
// 
// 		isStartJobCalled := make(chan bool, 1)
// 		gock.New(apiURL).
// 			Put("/jobs/" + childrenJob.UUID + "/start").
// 			AddMatcher(waitTimeoutMatcher(isStartJobCalled, job.InternalData.RetryInterval+time.Second)).
// 			Reply(http2.StatusAccepted)
// 
// 		err := s.env.ucs.PendingJobUseCase().Execute(ctx, job)
// 
// 		require.NoError(t, err)
// 		assert.True(t, <-isStartJobCalled, "missing update job call")
// 	})
// }
// 
// func searchTxMatcher(jobUUID string) gock.MatchFunc {
// 	return func(rw *http2.Request, _ *gock.Request) (bool, error) {
// 		qParentJobUUID := rw.URL.Query().Get("parent_job_uuid")
// 		if qParentJobUUID == jobUUID {
// 			return true, nil
// 		}
// 		return false, nil
// 	}
// }
// 
// func waitTimeoutMatcher(c chan bool, duration time.Duration) gock.MatchFunc {
// 	go func() {
// 		<-time.After(duration)
// 		c <- false
// 	}()
// 
// 	return func(rw *http2.Request, _ *gock.Request) (bool, error) {
// 		c <- true
// 		return true, nil
// 	}
// }
