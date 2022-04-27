// +build e2e

package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/quorum-key-manager/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type privTransactionsTestSuite struct {
	suite.Suite
	env        *Environment
	ctx        context.Context
	streamUUID string
	cancel     context.CancelFunc
	kafkaTopic string
}

func TestPrivateTransactions(t *testing.T) {
	s := new(privTransactionsTestSuite)
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

func (s *privTransactionsTestSuite) SetupSuite() {
	err := s.env.Start()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test suite has completed")
}

func (s *privTransactionsTestSuite) TearDownSuite() {
	err := s.env.Stop()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test teardown has completed")
}

func (s *privTransactionsTestSuite) TestPrivateTransactions_EEA() {
	newAcc, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
	require.NoError(s.T(), err)

	besuChain, _, err := s.env.createChainWithStream("chain-besu-"+common.RandString(5), s.env.TestData.Nodes.Besu[0].URLs, "")
	require.NoError(s.T(), err)

	ContractID := "Counter"
	_, err = s.env.Client.RegisterContract(s.ctx, &types.RegisterContractRequest{
		ABI:              s.env.Artifacts[ContractID].ABI,
		Bytecode:         s.env.Artifacts[ContractID].Bytecode,
		DeployedBytecode: s.env.Artifacts[ContractID].DeployedBytecode,
		Name:             ContractID,
	})
	require.NoError(s.T(), err)

	s.T().Run("as a user I want to send a private transactions and be notified when ready", func(t *testing.T) {
		txDeployReq, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: besuChain.Name,
			Params: types.DeployContractParams{
				PrivateFrom:  s.env.TestData.Nodes.Besu[0].PrivateAddress[0],
				PrivateFor:   []string{s.env.TestData.Nodes.Besu[0].PrivateAddress[0], s.env.TestData.Nodes.Besu[1].PrivateAddress[0]},
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
				Protocol:     entities.EEAChainType,
			},
		})
		require.NoError(t, err)

		txRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txDeployReq.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.Equal(t, txRes.Data.Job.Receipt.PrivateFrom, s.env.TestData.Nodes.Besu[0].PrivateAddress[0])
		assert.NotEmpty(t, txRes.Data.Job.Receipt.PrivateFor)
		assert.NotEmpty(t, txRes.Data.Job.Receipt.ContractAddress)
	})

	s.T().Run("when an user sends a private transaction with invalid private sender it is notified on tx-recover", func(t *testing.T) {
		txDeployReq, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: besuChain.Name,
			Params: types.DeployContractParams{
				PrivateFrom:  s.env.TestData.Nodes.Besu[1].PrivateAddress[0],
				PrivateFor:   []string{s.env.TestData.Nodes.Besu[0].PrivateAddress[0], s.env.TestData.Nodes.Besu[1].PrivateAddress[0]},
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
				Protocol:     entities.EEAChainType,
			},
		})
		require.NoError(t, err)

		txRes, err := s.env.ConsumerTracker.WaitForTxFailedNotification(s.ctx, txDeployReq.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.NotEmpty(t, txRes.Data.Error)
	})

	s.T().Run("when an user sends a private transaction with invalid protocol fail with expected error", func(t *testing.T) {
		_, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: besuChain.Name,
			Params: types.DeployContractParams{
				PrivateFrom:  s.env.TestData.Nodes.Besu[0].PrivateAddress[0],
				PrivateFor:   []string{s.env.TestData.Nodes.Besu[0].PrivateAddress[0], s.env.TestData.Nodes.Besu[1].PrivateAddress[0]},
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
			},
		})
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})
}

func (s *privTransactionsTestSuite) TestPrivateTransactions_GoQuorum() {
	newAcc, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
	require.NoError(s.T(), err)

	goQuorumChain, _, err := s.env.createChainWithStream("chain-go-quorum-"+common.RandString(5), s.env.TestData.Nodes.GoQuorum[0].URLs, s.env.TestData.Nodes.GoQuorum[0].PrivateTxManagerURL)
	require.NoError(s.T(), err)

	ContractID := "SimpleToken"
	_, err = s.env.Client.RegisterContract(s.ctx, &types.RegisterContractRequest{
		ABI:              s.env.Artifacts[ContractID].ABI,
		Bytecode:         s.env.Artifacts[ContractID].Bytecode,
		DeployedBytecode: s.env.Artifacts[ContractID].DeployedBytecode,
		Name:             ContractID,
	})
	require.NoError(s.T(), err)

	s.T().Run("as a user I want to send a private transactions and be notified when ready", func(t *testing.T) {
		txDeployReq, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: goQuorumChain.Name,
			Params: types.DeployContractParams{
				PrivateFrom:  s.env.TestData.Nodes.GoQuorum[0].PrivateAddress[0],
				PrivateFor:   []string{s.env.TestData.Nodes.GoQuorum[0].PrivateAddress[0], s.env.TestData.Nodes.GoQuorum[1].PrivateAddress[0]},
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
				Protocol:     entities.GoQuorumChainType,
			},
		})
		require.NoError(t, err)

		txRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txDeployReq.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.NotEmpty(t, txRes.Data.Job.Receipt.ContractAddress)
	})

	s.T().Run("as a user I want to send a private transactions using mandatoryFor and skipping PrivateFrom and be notified when ready", func(t *testing.T) {
		txDeployReq, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: goQuorumChain.Name,
			Params: types.DeployContractParams{
				PrivateFor:   []string{s.env.TestData.Nodes.GoQuorum[0].PrivateAddress[0], s.env.TestData.Nodes.GoQuorum[1].PrivateAddress[0]},
				MandatoryFor: []string{s.env.TestData.Nodes.GoQuorum[0].PrivateAddress[0]},
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
				PrivacyFlag:  2,
				Protocol:     entities.GoQuorumChainType,
			},
		})
		require.NoError(t, err)

		txRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txDeployReq.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.NotEmpty(t, txRes.Data.Job.Receipt.ContractAddress)
	})

	s.T().Run("when an user sends a private transaction with invalid private sender it is notified", func(t *testing.T) {
		txDeployReq, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: goQuorumChain.Name,
			Params: types.DeployContractParams{
				PrivateFrom:  s.env.TestData.Nodes.GoQuorum[1].PrivateAddress[0],
				PrivateFor:   []string{s.env.TestData.Nodes.GoQuorum[0].PrivateAddress[0], s.env.TestData.Nodes.GoQuorum[1].PrivateAddress[0]},
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
				Protocol:     entities.GoQuorumChainType,
			},
		})
		require.NoError(t, err)

		txRes, err := s.env.ConsumerTracker.WaitForTxFailedNotification(s.ctx, txDeployReq.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.NotEmpty(t, txRes.Data.Error)
	})

	s.T().Run("when an user sends a private transaction with invalid protocol fail with expected error", func(t *testing.T) {
		_, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: goQuorumChain.Name,
			Params: types.DeployContractParams{
				PrivateFrom:  s.env.TestData.Nodes.GoQuorum[1].PrivateAddress[0],
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
				Protocol:     entities.GoQuorumChainType,
			},
		})
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})
}
