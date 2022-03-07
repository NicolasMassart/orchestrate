// +build integration

package integrationtests

import (
	"fmt"
	http2 "net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	types2 "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/infra/ethclient/types"
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

	s.T().Run("should get pending job and wait for mined transaction successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accOneTxHashOne)).(*ethcommon.Hash)

		gock.New(apiURL).
			Patch("/jobs/" + job.UUID).
			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))

		err := s.env.ucs.PendingJobUseCase().Execute(ctx, job)
		require.NoError(t, err)

		hash, err := s.env.ethClient.SendRawTransaction(ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accOneRawTxOne))
		require.NoError(t, err)

		assert.Equal(t, hash.String(), job.Transaction.Hash.String())

		time.Sleep(s.env.chain.ListenerBackOffDuration + time.Second)

		evlp, err := s.env.consumer.WaitForEnvelope(job.ScheduleUUID, s.env.cfg.ChainListenerConfig.DecodedOutTopic,
			waitForEnvelopeTimeOut)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		assert.Equal(t, evlp.GetJobUUID(), job.UUID)
	})

	s.T().Run("should get pending job and fetch available receipt right away successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID
		job.Transaction.Hash = utils.ToPtr(ethcommon.HexToHash(accOneTxHashTwo)).(*ethcommon.Hash)

		gock.New(apiURL).
			Patch("/jobs/" + job.UUID).
			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))

		hash, err := s.env.ethClient.SendRawTransaction(ctx, s.env.blockchainNodeURL, hexutil.MustDecode(accOneRawTxTwo))
		require.NoError(t, err)

		assert.Equal(t, hash.String(), job.Transaction.Hash.String())

		time.Sleep(s.env.chain.ListenerBackOffDuration + time.Second)

		err = s.env.ucs.PendingJobUseCase().Execute(ctx, job)
		require.NoError(t, err)

		time.Sleep(time.Second)

		evlp, err := s.env.consumer.WaitForEnvelope(job.ScheduleUUID, s.env.cfg.ChainListenerConfig.DecodedOutTopic,
			waitForEnvelopeTimeOut)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		assert.Equal(t, evlp.GetJobUUID(), job.UUID)
	})

	s.T().Run("should get pending job, fetch receipt and register contract successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.ChainUUID = s.env.chain.UUID

		contractName := "Counter"
		contractTag := "latest"
		
		gock.New(apiURL).
			Patch("/jobs/" + job.UUID).
			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))

		hash, err := s.env.ethClient.SendTransaction(ctx, s.env.blockchainNodeURL, &types.SendTxArgs{
			From: ethcommon.HexToAddress("0xdbb881a51CD4023E4400CEF3ef73046743f08da3"),
			Data: hexutil.MustDecode("0x608060405234801561001057600080fd5b5061023c806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80637cf5dab014610030575b600080fd5b61004a600480360381019061004591906100db565b61004c565b005b8060008082825461005d9190610137565b925050819055507f38ac789ed44572701765277c4d0970f2db1c1a571ed39e84358095ae4eaa542033826040516100959291906101dd565b60405180910390a150565b600080fd5b6000819050919050565b6100b8816100a5565b81146100c357600080fd5b50565b6000813590506100d5816100af565b92915050565b6000602082840312156100f1576100f06100a0565b5b60006100ff848285016100c6565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610142826100a5565b915061014d836100a5565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561018257610181610108565b5b828201905092915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006101b88261018d565b9050919050565b6101c8816101ad565b82525050565b6101d7816100a5565b82525050565b60006040820190506101f260008301856101bf565b6101ff60208301846101ce565b939250505056fea264697066735822122072834524b3ee30e2b953db63515fc66272b7245946cfc2523dcc3e81b659ac6464736f6c63430008090033"),
		})
		require.NoError(t, err)
		job.Transaction.Hash = &hash

		time.Sleep(time.Second)
		
		receipt, err := s.env.ethClient.TransactionReceipt(ctx, s.env.blockchainNodeURL, hash)
		require.NoError(t, err)
		
		gock.New(apiURL).
			Post(fmt.Sprintf("contracts/accounts/%s/%s", s.env.chain.ChainID, receipt.ContractAddress)).
			Reply(http2.StatusOK).JSON(formatters.FormatJobResponse(job))
		
		gock.New(apiURL).
			Get("contracts/search").
			Reply(http2.StatusOK).JSON(&types2.ContractResponse{
				Name: contractName,
				Tag: contractTag,
		})
		
		err = s.env.ucs.PendingJobUseCase().Execute(ctx, job)
		require.NoError(t, err)

		time.Sleep(s.env.chain.ListenerBackOffDuration + time.Second)

		evlp, err := s.env.consumer.WaitForEnvelope(job.ScheduleUUID, s.env.cfg.ChainListenerConfig.DecodedOutTopic,
			waitForEnvelopeTimeOut)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		assert.Equal(t, evlp.GetJobUUID(), job.UUID)
		assert.Equal(t, evlp.GetContractName(), contractName)
		assert.Equal(t, evlp.GetContractTag(), contractTag)
	})
}
