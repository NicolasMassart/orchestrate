// +build e2e

package e2e

import (
	"context"
	"math/big"
	"os"
	"testing"

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

type nonceManagementTestSuite struct {
	suite.Suite
	env        *Environment
	ctx        context.Context
	cancel     context.CancelFunc
	kafkaTopic string
}

func TestNonceManagement(t *testing.T) {
	s := new(nonceManagementTestSuite)
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

func (s *nonceManagementTestSuite) SetupSuite() {
	err := s.env.Start()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test suite has completed")
}

func (s *nonceManagementTestSuite) TearDownSuite() {
	err := s.env.Stop()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test teardown has completed")
}

func (s *nonceManagementTestSuite) TestNonceManagement_Recalibrate() {
	faucetAcc, err := pkgutils.ImportOrFetchAccount(s.ctx, s.env.Client, s.env.TestData.Nodes.Geth[0].FundedPublicKeys[0], &types.ImportAccountRequest{
		Alias:      "faucet-acc-" + common.RandString(5),
		PrivateKey: s.env.TestData.Nodes.Geth[0].FundedPrivateKeys[0],
	})
	require.NoError(s.T(), err)

	gethChain, _, err := s.env.createChainWithStream("chain-geth-"+common.RandString(5), s.env.TestData.Nodes.Geth[0].URLs, "")
	require.NoError(s.T(), err)

	s.T().Run("as a user I want to send a batch of transactions and recover from failed one in sequence", func(t *testing.T) {
		testAcc, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
		require.NoError(s.T(), err)

		faucetTxRes, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(faucetAcc.Address),
				To:    ethcommon.HexToAddress(testAcc.Address),
				Value: utils.HexToBigInt("0xDE0B6B3A7640000"), // 1ETH
			},
		})
		require.NoError(s.T(), err)
		_, err = s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, faucetTxRes.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)

		txOne, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(testAcc.Address),
				To:    ethcommon.HexToAddress(faucetAcc.Address),
				Value: utils.HexToBigInt("0x16345785D8A0000"), // 0.1ETH
			},
		})
		require.NoError(s.T(), err)

		// Fail due to not enough balance
		txTwo, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(testAcc.Address),
				To:    ethcommon.HexToAddress(faucetAcc.Address),
				Value: utils.HexToBigInt("0xDE0B6B3A7640000"), // 1 ETH
			},
		})

		txThree, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(testAcc.Address),
				To:    ethcommon.HexToAddress(faucetAcc.Address),
				Value: utils.HexToBigInt("0x16345785D8A0000"), // 0.1ETH
			},
		})
		require.NoError(s.T(), err)

		txOneRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txOne.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)
		assert.Equal(t, uint64(0), *txOneRes.Data.Job.Transaction.Nonce)

		require.NoError(s.T(), err)
		txTwoRes, err := s.env.ConsumerTracker.WaitForTxFailedNotification(s.ctx, txTwo.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)
		assert.NotEmpty(t, txTwoRes.Data.Error)

		txThreeRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txThree.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)
		assert.Equal(t, uint64(1), *txThreeRes.Data.Job.Transaction.Nonce)
	})

	s.T().Run("as a user I want to recover nonce in case of using same account externally", func(t *testing.T) {
		// Create and import a new account
		testAccPrivKey, err := ethcrypto.GenerateKey()
		require.NoError(s.T(), err)
		testAccAddr := ethcrypto.PubkeyToAddress(testAccPrivKey.PublicKey)
		testAcc, err := pkgutils.ImportOrFetchAccount(s.ctx, s.env.Client, testAccAddr.String(), &types.ImportAccountRequest{
			Alias:      "test-nonce-acc-" + common.RandString(5),
			PrivateKey: ethcrypto.FromECDSA(testAccPrivKey),
		})
		require.NoError(s.T(), err)
		require.Equal(s.T(), testAccAddr.String(), testAcc.Address)

		// Fund the new account
		faucetTxRes, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(faucetAcc.Address),
				To:    ethcommon.HexToAddress(testAcc.Address),
				Value: utils.HexToBigInt("0xDE0B6B3A7640000"), // 1ETH
				GasPricePolicy: types.GasPriceParams{
					Priority: entities.GasIncrementVeryHigh,
				},
			},
		})
		require.NoError(s.T(), err)
		_, err = s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, faucetTxRes.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)

		// Send tx with nonce 0
		txOne, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(testAcc.Address),
				To:    ethcommon.HexToAddress(faucetAcc.Address),
				Value: utils.HexToBigInt("0x16345785D8A0000"), // 0.1ETH
			},
		})
		require.NoError(s.T(), err)

		// Send External TX with nonce 1
		gasPrice, err := s.env.EthClient.SuggestGasPrice(s.ctx, s.env.Client.ChainProxyURL(gethChain.UUID))
		require.NoError(s.T(), err)
		signedRaw, _, err := crypto.SignTransaction(testAccPrivKey, &entities.ETHTransaction{
			From:            utils.HexToAddress(testAcc.Address),
			To:              utils.HexToAddress(faucetAcc.Address),
			Nonce:           utils.ToPtr(uint64(1)).(*uint64),
			TransactionType: entities.LegacyTxType,
			GasPrice:        (*hexutil.Big)(new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(2))),
			Gas:             utils.ToPtr(uint64(210000)).(*uint64),
			Value:           utils.HexToBigInt("0x16345785D8A0000"), // 0.1ETH
		}, gethChain.ChainID)
		require.NoError(s.T(), err)

		txTwo, err := s.env.Client.SendRawTransaction(s.ctx, &types.RawTransactionRequest{
			ChainName: gethChain.Name,
			Params: types.RawTransactionParams{
				Raw: signedRaw,
			},
		})

		txOneRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, s.env.KafkaTopic, txOne.UUID, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)
		assert.Equal(t, uint64(0), *txOneRes.Data.Job.Transaction.Nonce)

		_, err = s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txTwo.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)

		// Send tx with recalibrated to nonce 2
		txThree, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(testAcc.Address),
				To:    ethcommon.HexToAddress(faucetAcc.Address),
				Value: utils.HexToBigInt("0x16345785D8A0000"), // 0.1ETH
				GasPricePolicy: types.GasPriceParams{
					Priority: entities.GasIncrementVeryHigh,
				},
			},
		})
		require.NoError(s.T(), err)
		txThreeRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txThree.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)
		assert.Equal(t, uint64(2), *txThreeRes.Data.Job.Transaction.Nonce)
	})
}
