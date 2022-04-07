// +build e2e

package e2e

import (
	"context"
	"math/big"
	"net/http"
	"os"
	"testing"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/tests/pkg/crypto"
	pkgutils "github.com/consensys/orchestrate/tests/pkg/utils"
	"github.com/consensys/quorum-key-manager/pkg/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type transactionTestSuite struct {
	suite.Suite
	env    *Environment
	ctx    context.Context
	cancel context.CancelFunc
}

func TestTransactions(t *testing.T) {
	s := new(transactionTestSuite)
	s.ctx, s.cancel = context.WithCancel(context.Background())

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

func (s *transactionTestSuite) TestTransferTransactions() {
	sourceAcc, err := pkgutils.ImportOrFetchAccount(s.ctx, s.env.Client, s.env.TestData.Nodes.Geth[0].FundedPublicKeys[0], &types.ImportAccountRequest{
		Alias:      "source-acc-" + common.RandString(5),
		PrivateKey: s.env.TestData.Nodes.Geth[0].FundedPrivateKeys[0],
	})
	require.NoError(s.T(), err)

	chainRes, err := pkgutils.RegisterChainAndWaitForProxy(s.ctx, s.env.Client, s.env.EthClient, &types.RegisterChainRequest{
		Name: "chain-geth-" + common.RandString(5),
		URLs: s.env.TestData.Nodes.Geth[0].URLs,
	})

	require.NoError(s.T(), err)
	defer func() {
		err := s.env.Client.DeleteChain(s.ctx, chainRes.UUID)
		require.NoError(s.T(), err)
	}()

	s.T().Run("as a user I want to send a transaction and be notified when ready", func(t *testing.T) {
		newAccRes, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
		require.NoError(s.T(), err)

		transferParams := types.TransferParams{
			From:  ethcommon.HexToAddress(sourceAcc.Address),
			To:    ethcommon.HexToAddress(newAccRes.Address),
			Value: utils.HexToBigInt("0x16345785D8A0000"),
		}
		transactionRes, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: chainRes.Name,
			Params:    transferParams,
		})
		require.NoError(t, err)

		txRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, transactionRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.Equal(t, txRes.Data.Job.Transaction.From, sourceAcc.Address)
		assert.Equal(t, txRes.Data.Job.Transaction.To, newAccRes.Address)

		balance, err := s.env.EthClient.BalanceAt(s.ctx, s.env.Client.ChainProxyURL(chainRes.UUID), ethcommon.HexToAddress(newAccRes.Address), nil)
		require.NoError(t, err)
		assert.Equal(t, transferParams.Value.String(), hexutil.EncodeBig(balance))
	})

	s.T().Run("as a user I want to send a transaction and be notified if it fails", func(t *testing.T) {
		newAccRes, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
		require.NoError(s.T(), err)

		transferParams := types.TransferParams{
			From:  ethcommon.HexToAddress(newAccRes.Address),
			To:    ethcommon.HexToAddress(sourceAcc.Address),
			Value: utils.HexToBigInt("0x16345785D8A0000"),
		}
		transactionRes, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: chainRes.Name,
			Params:    transferParams,
		})

		require.NoError(t, err)

		txRes, err := s.env.KafkaConsumer.WaitForTxFailedNotification(s.ctx, transactionRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		require.NotNil(t, txRes)
	})

	s.T().Run("when an user sends a transfer transaction with invalid parameters fail with expected error", func(t *testing.T) {
		newAccRes, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
		require.NoError(s.T(), err)

		transferParams := types.TransferParams{
			From: ethcommon.HexToAddress(newAccRes.Address),
			To:   ethcommon.HexToAddress(sourceAcc.Address),
		}
		_, err = s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: chainRes.Name,
			Params:    transferParams,
		})

		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})
}

func (s *transactionTestSuite) TestContractTransactions() {
	chainRes, err := pkgutils.RegisterChainAndWaitForProxy(s.ctx, s.env.Client, s.env.EthClient, &types.RegisterChainRequest{
		Name: "chain-besu-" + common.RandString(5),
		URLs: s.env.TestData.Nodes.Besu[0].URLs,
	})

	require.NoError(s.T(), err)
	defer func() {
		err := s.env.Client.DeleteChain(s.ctx, chainRes.UUID)
		require.NoError(s.T(), err)
	}()

	newAcc, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
	require.NoError(s.T(), err)
	txLabels := map[string]string{
		"labelOne": "valueOne",
	}

	s.T().Run("as a user I want to deploy a Counter smart contract and increment its counter", func(t *testing.T) {
		ContractID := "Counter"
		_, err = s.env.Client.RegisterContract(s.ctx, &types.RegisterContractRequest{
			ABI:              s.env.Artifacts[ContractID].ABI,
			Bytecode:         s.env.Artifacts[ContractID].Bytecode,
			DeployedBytecode: s.env.Artifacts[ContractID].DeployedBytecode,
			Name:             ContractID,
		})
		require.NoError(s.T(), err)

		txRes, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: chainRes.Name,
			Labels:    txLabels,
			Params: types.DeployContractParams{
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
			},
		})
		require.NoError(t, err)

		deployTxRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		require.Equal(t, uint64(1), deployTxRes.Data.Job.Receipt.Status)

		txRes, err = s.env.Client.SendContractTransaction(s.ctx, &types.SendTransactionRequest{
			ChainName: chainRes.Name,
			Labels:    txLabels,
			Params: types.TransactionParams{
				From:            utils.HexToAddress(newAcc.Address),
				To:              utils.HexToAddress(deployTxRes.Data.Job.Receipt.ContractAddress),
				MethodSignature: "increment(uint256)",
				Args:            []interface{}{2},
				ContractName:    ContractID,
			},
		})
		require.NoError(t, err)

		contractTxRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), contractTxRes.Data.Job.Receipt.Status)
		assert.NotEmpty(t, contractTxRes.Data.Job.Receipt.Logs)
	})

	s.T().Run("as a user I want to deploy a SimpleToken smart contract and send a token transfer", func(t *testing.T) {
		ContractID := "SimpleToken"
		_, err = s.env.Client.RegisterContract(s.ctx, &types.RegisterContractRequest{
			ABI:              s.env.Artifacts[ContractID].ABI,
			Bytecode:         s.env.Artifacts[ContractID].Bytecode,
			DeployedBytecode: s.env.Artifacts[ContractID].DeployedBytecode,
			Name:             ContractID,
		})
		require.NoError(s.T(), err)

		destAcc, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
		require.NoError(s.T(), err)

		txRes, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: chainRes.Name,
			Labels:    txLabels,
			Params: types.DeployContractParams{
				From:         utils.HexToAddress(newAcc.Address),
				ContractName: ContractID,
			},
		})
		require.NoError(t, err)

		deployTxRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		require.Equal(t, uint64(1), deployTxRes.Data.Job.Receipt.Status)

		txRes, err = s.env.Client.SendContractTransaction(s.ctx, &types.SendTransactionRequest{
			ChainName: chainRes.Name,
			Labels:    txLabels,
			Params: types.TransactionParams{
				From:            utils.HexToAddress(newAcc.Address),
				To:              utils.HexToAddress(deployTxRes.Data.Job.Receipt.ContractAddress),
				MethodSignature: "transfer(address,uint256)",
				Args:            []interface{}{destAcc.Address, 1},
				ContractName:    ContractID,
			},
		})
		require.NoError(t, err)

		contractTxRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), contractTxRes.Data.Job.Receipt.Status)
		require.NotEmpty(t, contractTxRes.Data.Job.Receipt.Logs)
		assert.Equal(t, newAcc.Address, contractTxRes.Data.Job.Receipt.Logs[0].DecodedData["from"])
		assert.Equal(t, destAcc.Address, contractTxRes.Data.Job.Receipt.Logs[0].DecodedData["to"])
	})

	s.T().Run("as a user I want to deploy a smart contract using latest supported types using one time keys", func(t *testing.T) {
		contractID := "Struct"
		_, err = s.env.Client.RegisterContract(s.ctx, &types.RegisterContractRequest{
			ABI:              s.env.Artifacts[contractID].ABI,
			Bytecode:         s.env.Artifacts[contractID].Bytecode,
			DeployedBytecode: s.env.Artifacts[contractID].DeployedBytecode,
			Name:             contractID,
		})
		require.NoError(s.T(), err)

		txRes, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: chainRes.Name,
			Labels:    txLabels,
			Params: types.DeployContractParams{
				OneTimeKey:   true,
				ContractName: contractID,
			},
		})
		require.NoError(t, err)

		deployTxRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), deployTxRes.Data.Job.Receipt.Status)
	})

	s.T().Run("as a user I want to deploy a ERC20 smart contract", func(t *testing.T) {
		ContractID := "ERC20"
		_, err = s.env.Client.RegisterContract(s.ctx, &types.RegisterContractRequest{
			ABI:              s.env.Artifacts[ContractID].ABI,
			Bytecode:         s.env.Artifacts[ContractID].Bytecode,
			DeployedBytecode: s.env.Artifacts[ContractID].DeployedBytecode,
			Name:             ContractID,
		})
		require.NoError(s.T(), err)

		txRes, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: chainRes.Name,
			Labels:    txLabels,
			Params: types.DeployContractParams{
				From:         utils.HexToAddress(newAcc.Address),
				Args:         []interface{}{"token_name", "token_symbol"},
				ContractName: ContractID,
			},
		})
		require.NoError(t, err)

		deployTxRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		require.Equal(t, uint64(1), deployTxRes.Data.Job.Receipt.Status)
		require.NotEmpty(t, deployTxRes.Data.Job.Receipt.ContractAddress)
	})

	s.T().Run("when an user to try deploy contract with invalid arguments if should fail with expected error", func(t *testing.T) {
		ContractID := "ERC20"
		_, err = s.env.Client.RegisterContract(s.ctx, &types.RegisterContractRequest{
			ABI:              s.env.Artifacts[ContractID].ABI,
			Bytecode:         s.env.Artifacts[ContractID].Bytecode,
			DeployedBytecode: s.env.Artifacts[ContractID].DeployedBytecode,
			Name:             ContractID,
		})
		require.NoError(s.T(), err)

		_, err := s.env.Client.SendDeployTransaction(s.ctx, &types.DeployContractRequest{
			ChainName: chainRes.Name,
			Labels:    txLabels,
			Params: types.DeployContractParams{
				From:         utils.HexToAddress(newAcc.Address),
				Args:         []interface{}{1, "symbol"},
				ContractName: ContractID,
			},
		})
		require.Error(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, err.(*client.HTTPErr).Code())
	})
}

func (s *transactionTestSuite) TestRawTransactions() {
	chainRes, err := pkgutils.RegisterChainAndWaitForProxy(s.ctx, s.env.Client, s.env.EthClient, &types.RegisterChainRequest{
		Name: "chain-go-quorum-" + common.RandString(5),
		URLs: s.env.TestData.Nodes.GoQuorum[0].URLs,
	})
	require.NoError(s.T(), err)
	chainID := new(big.Int).SetUint64(chainRes.ChainID)
	defer func() {
		err := s.env.Client.DeleteChain(s.ctx, chainRes.UUID)
		require.NoError(s.T(), err)
	}()

	gasPrice, err := s.env.EthClient.SuggestGasPrice(s.ctx, s.env.Client.ChainProxyURL(chainRes.UUID))
	require.NoError(s.T(), err)

	privKey, err := ethcrypto.GenerateKey()
	require.NoError(s.T(), err)
	accAddr := ethcrypto.PubkeyToAddress(privKey.PublicKey)

	s.T().Run("as a user I want to send a raw transaction and be notified once it is mined", func(t *testing.T) {
		ethTransaction := &entities.ETHTransaction{
			From:            &accAddr,
			Nonce:           utils.ToPtr(uint64(0)).(*uint64),
			TransactionType: entities.LegacyTxType,
			GasPrice:        (*hexutil.Big)(gasPrice),
			Gas:             utils.ToPtr(uint64(210000)).(*uint64),
			Data:            hexutil.MustDecode("0x"),
		}

		signedRaw, _, err := crypto.SignTransaction(privKey, ethTransaction, chainID)
		require.NoError(s.T(), err)

		txRes, err := s.env.Client.SendRawTransaction(s.ctx, &types.RawTransactionRequest{
			ChainName: chainRes.Name,
			Params: types.RawTransactionParams{
				Raw: signedRaw,
			},
		})

		require.NoError(s.T(), err)
		deployTxRes, err := s.env.KafkaConsumer.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.TestData.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(t, err)
		require.Equal(t, uint64(1), deployTxRes.Data.Job.Receipt.Status)
	})
	
	s.T().Run("when an user sends a raw transaction with invalid arguments", func(t *testing.T) {
		require.NoError(s.T(), err)

		_, err := s.env.Client.SendRawTransaction(s.ctx, &types.RawTransactionRequest{
			ChainName: chainRes.Name,
			Params: types.RawTransactionParams{
				Raw: nil,
			},
		})

		require.Error(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, err.(*client.HTTPErr).Code())
	})
}
