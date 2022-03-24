// +build integration

package integrationtests

import (
	http2 "net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	types2 "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	infra "github.com/consensys/orchestrate/src/infra/api"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type txListenerTestSuite struct {
	suite.Suite
	env *IntegrationEnvironment
}

// From: 0xdbb881a51CD4023E4400CEF3ef73046743f08da3
const accOneRawTxOne = "0xf86e808443a5b73882520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820713a0c23154893b531af779316139fd21a59643ec446209bda55afb23ba9cc1ec5720a00355d30b02fe453f3e9595a4b7fbf7a290560e3e1d4f5869167ed5870e8d69a7"
const accOneTxHashOne = "0x4c280d7f7b9ae00b44b472156a898e19ee96ec07b89d770dc124734306b431ea"

const accOneRawTxTwo = "0xf86e018442a4599182520894697f7da20b3c84dfe38ed933ab6dbc81ee25d70e89056bc75e2d6310000080820713a08193ae278c6f7bb0568105873ba899266110d9d8f7aca1bd5300554bd6a18a6da00d96cf498d9f2649bfdd1daecb8fe29b6a647a4cf1efde0df21d697f336a1683"
const accOneTxHashTwo = "0xb0ee44a2d19c554790ff89de85230284e2956678e4ce4967a3f237b8fa807fb2"

func (s *txListenerTestSuite) TestNotifyMinedTx() {
	ctx := s.env.ctx
	chain := testdata.FakeChain()
	chain.UUID = ganacheChainUUID

	respWaitingTime := s.env.chain.ListenerBackOffDuration + time.Second
	s.T().Run("should get pending job and wait for mined transaction successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accOneTxHashOne)).(*ethcommon.Hash)

		isExpectedCall := make(chan bool, 1)
		gock.New(apiURL).
			AddMatcher(statusJobUpdateMatcher(isExpectedCall, respWaitingTime, func(req *types2.UpdateJobRequest) error {
				assert.Equal(t, entities.StatusMined, req.Status)
				assert.Equal(t, accOneTxHashOne, req.Receipt.TxHash)
				return nil
			})).
			Patch("/jobs/" + job.UUID).
			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))

		err := s.env.ucs.PendingJobUseCase().Execute(ctx, job)
		require.NoError(t, err)

		hash, err := s.env.ethClient.SendRawTransaction(ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accOneRawTxOne))
		require.NoError(t, err)

		assert.Equal(t, hash.String(), job.Transaction.Hash.String())

		assert.True(t, <-isExpectedCall)
	})

	s.T().Run("should get pending job and fetch available receipt right away successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accOneTxHashTwo)).(*ethcommon.Hash)

		isExpectedCall := make(chan bool, 1)
		gock.New(apiURL).
			Patch("/jobs/" + job.UUID).
			AddMatcher(statusJobUpdateMatcher(isExpectedCall, respWaitingTime, func(req *types2.UpdateJobRequest) error {
				assert.Equal(t, entities.StatusMined, req.Status)
				assert.Equal(t, accOneTxHashTwo, req.Receipt.TxHash)
				return nil
			})).
			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))

		hash, err := s.env.ethClient.SendRawTransaction(ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accOneRawTxTwo))
		require.NoError(t, err)

		assert.Equal(t, hash.String(), job.Transaction.Hash.String())

		err = s.env.ucs.PendingJobUseCase().Execute(ctx, job)

		require.NoError(t, err)
		assert.True(t, <-isExpectedCall)
	})
}

func statusJobUpdateMatcher(c chan bool, waiting time.Duration, cb func(req *types2.UpdateJobRequest) error) gock.MatchFunc {
	go func() {
		time.Sleep(waiting)
		c <- false
	}()

	return func(rw *http2.Request, req *gock.Request) (bool, error) {
		c <- true
		jobRequest := &types2.UpdateJobRequest{}
		err := infra.UnmarshalBody(rw.Body, jobRequest)
		if err != nil {
			return false, err
		}

		return true, cb(jobRequest)
	}
}
