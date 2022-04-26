// +build e2e

package e2e

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

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

type txSentryTestSuite struct {
	suite.Suite
	env        *Environment
	ctx        context.Context
	cancel     context.CancelFunc
	kafkaTopic string
}

func TestTxSentry(t *testing.T) {
	s := new(txSentryTestSuite)
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

func (s *txSentryTestSuite) SetupSuite() {
	err := s.env.Start()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test suite has completed")
}

func (s *txSentryTestSuite) TearDownSuite() {
	err := s.env.Stop()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test teardown has completed")
}

func (s *txSentryTestSuite) TestTxSentry_Successful() {
	faucetAcc, err := pkgutils.ImportOrFetchAccount(s.ctx, s.env.Client, s.env.TestData.Nodes.Geth[0].FundedPublicKeys[0], &types.ImportAccountRequest{
		Alias:      "faucet-geth-" + common.RandString(5),
		PrivateKey: s.env.TestData.Nodes.Geth[0].FundedPrivateKeys[0],
	})
	require.NoError(s.T(), err)

	gethChain, _, err := s.env.createChainWithStream("chain-geth-"+common.RandString(5), s.env.TestData.Nodes.Geth[0].URLs, "")
	require.NoError(s.T(), err)

	spammerAcc, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
	require.NoError(s.T(), err)

	faucetSpammerTxRes, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
		ChainName: gethChain.Name,
		Params: types.TransferParams{
			From:  ethcommon.HexToAddress(faucetAcc.Address),
			To:    ethcommon.HexToAddress(spammerAcc.Address),
			Value: utils.HexToBigInt("0x8AC7230489E80000"), // 10ETH
		},
	})
	require.NoError(s.T(), err)
	_, err = s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, faucetSpammerTxRes.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
	require.NoError(s.T(), err)

	// Run in loop to simulate an active network
	var endLoop bool
	defer func() {
		endLoop = true
	}()
	go func() {
		for {
			if endLoop {
				break
			}
			spammerTx, err := s.env.Client.SendTransferTransaction(s.ctx, &types.TransferRequest{
				ChainName: gethChain.Name,
				Params: types.TransferParams{
					From:  ethcommon.HexToAddress(spammerAcc.Address),
					To:    ethcommon.HexToAddress(faucetAcc.Address),
					Value: utils.HexToBigInt("0x16345785D8A0000"), // 0.1ETH
				},
			})
			require.NoError(s.T(), err)
			_, err = s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, spammerTx.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
			assert.NoError(s.T(), err)
		}
	}()
	// //

	s.T().Run("as a user I want to send a transaction with low priority and retry with a predefine gas increment", func(t *testing.T) {
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
				GasPricePolicy: types.GasPriceParams{
					Priority: entities.PriorityVeryLow,
					RetryPolicy: types.RetryParams{
						Interval:  "2s",
						Increment: 0.2,
						Limit:     1.2,
					},
				},
			},
		})
		require.NoError(s.T(), err)

		txOneRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txOne.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)
		require.Equal(t, uint64(1), txOneRes.Data.Job.Receipt.Status)

		finalTxReqStatus, err := s.env.Client.GetTxRequest(s.ctx, txOne.UUID)
		require.NoError(s.T(), err)
		assert.GreaterOrEqual(t, len(finalTxReqStatus.Jobs), 3)

		// Only one job should be mined and rest never mined
		minedJobFound := false
		var prevGasTipCap *big.Int
		for _, job := range finalTxReqStatus.Jobs {
			gasTipCap, _ := new(big.Int).SetString(job.Transaction.GasTipCap, 10)
			if prevGasTipCap != nil {
				assert.Greater(t, gasTipCap.Uint64(), prevGasTipCap.Uint64())
			}
			if !minedJobFound && job.Status == entities.StatusMined {
				minedJobFound = true
			} else {
				assert.Equal(t, entities.StatusNeverMined, job.Status)
			}

			prevGasTipCap = gasTipCap
		}
		assert.True(t, minedJobFound, "not mined job found")
	})

	s.T().Run("as a user I want to send a transaction with medium priority and no gas increment", func(t *testing.T) {
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
				GasPricePolicy: types.GasPriceParams{
					Priority: entities.PriorityMedium,
					RetryPolicy: types.RetryParams{
						Interval: "1s",
					},
				},
			},
		})
		require.NoError(s.T(), err)

		txOneRes, err := s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txOne.UUID, s.env.KafkaTopic, s.env.WaitForTxResponseTTL)
		require.NoError(s.T(), err)
		require.Equal(t, uint64(1), txOneRes.Data.Job.Receipt.Status)

		finalTxReqStatus, err := s.env.Client.GetTxRequest(s.ctx, txOne.UUID)
		require.NoError(s.T(), err)
		assert.Greater(t, len(finalTxReqStatus.Jobs), 2)

		// Only one job should be mined and rest never mined
		minedJobFound := false
		var prevGasTipCap *big.Int
		for _, job := range finalTxReqStatus.Jobs {
			gasTipCap, _ := new(big.Int).SetString(job.Transaction.GasTipCap, 10)
			if prevGasTipCap != nil {
				assert.Equal(t, gasTipCap.Uint64(), prevGasTipCap.Uint64())
			}
			if !minedJobFound && job.Status == entities.StatusMined {
				minedJobFound = true
			} else {
				assert.Equal(t, entities.StatusNeverMined, job.Status)
			}

			prevGasTipCap = gasTipCap
		}
		assert.True(t, minedJobFound, "not mined job found")
	})
}

func (s *txSentryTestSuite) TestTxSentry_Failure() {
	chainRes, err := pkgutils.RegisterChainAndWaitForProxy(s.ctx, s.env.Client, s.env.EthClient, &types.RegisterChainRequest{
		Name: "chain-besu-" + common.RandString(5),
		URLs: s.env.TestData.Nodes.Besu[0].URLs,
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

	s.T().Run("as a user I want to send a raw transaction and retry it till it reaches max retries limit (10 times)", func(t *testing.T) {
		ethTransaction := &entities.ETHTransaction{
			From:            &accAddr,
			Nonce:           utils.ToPtr(uint64(1)).(*uint64),
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
				RetryPolicy: types.IntervalRetryParams{
					Interval: "1s",
				},
			},
		})

		require.NoError(s.T(), err)
		_, err = s.env.ConsumerTracker.WaitForTxMinedNotification(s.ctx, txRes.UUID, s.env.KafkaTopic, time.Second*(types.SentryMaxRetries+5))
		require.Error(t, err)

		finalTxReqStatus, err := s.env.Client.GetTxRequest(s.ctx, txRes.UUID)
		require.NoError(s.T(), err)
		assert.Equal(t, len(finalTxReqStatus.Jobs), 1)
		assert.True(t, finalTxReqStatus.Jobs[0].Annotations.HasBeenRetried)
	})
}
