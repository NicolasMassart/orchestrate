// +build unit

package sender

import (
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/handlers/sender/mocks"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/engine"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/ethereum/types"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/entities"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/tx"
	txschedulertypes "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/txscheduler"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/utils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/proxy"
	mock2 "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/client/mock"
)

const chainRegistryUrl = "chainRegistryUrl"

type updateStatusMatcher struct {
	x *txschedulertypes.UpdateJobRequest
}

func gomockUpdateStatusMatcher(x *txschedulertypes.UpdateJobRequest) updateStatusMatcher {
	return updateStatusMatcher{
		x: x,
	}
}

func (e updateStatusMatcher) Matches(x interface{}) bool {
	if xt, ok := x.(*txschedulertypes.UpdateJobRequest); ok {
		return e.x.Status == xt.Status
	}
	return false
}

func (e updateStatusMatcher) String() string {
	return e.x.Status
}

func newTxCtx(eId, txHash, txRaw string) *engine.TxContext {
	txctx := engine.NewTxContext()
	txctx.Logger = log.NewEntry(log.StandardLogger())
	txctx.WithContext(proxy.With(txctx.Context(), chainRegistryUrl))
	_ = txctx.Envelope.SetID(eId).SetJobUUID(eId).SetTxHash(ethcommon.HexToHash(txHash)).SetRawString(txRaw)

	return txctx
}

func TestSender_RawTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	envelopeId := utils.RandomString(12)
	txHash := "0x" + utils.RandHexString(64)
	txRaw := "0x" + utils.RandHexString(10)

	schedulerClient := mock2.NewMockTransactionSchedulerClient(ctrl)

	ec := mocks.NewMockEthClient(ctrl)
	sender := Sender(ec, schedulerClient)

	t.Run("should execute raw transaction", func(t *testing.T) {
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.SetJobType(tx.JobType_ETH_RAW_TX)

		ec.EXPECT().SendRawTransaction(txctx.Context(), chainRegistryUrl, txRaw).Return(ethcommon.HexToHash(txHash), nil)

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))

		sender(txctx)
	})

	t.Run("should fail execute raw transaction", func(t *testing.T) {
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.SetJobType(tx.JobType_ETH_RAW_TX)
		err := fmt.Errorf("failed to send a raw transaction")

		ec.EXPECT().SendRawTransaction(txctx.Context(), chainRegistryUrl, txRaw).
			Return(ethcommon.Hash{}, err)

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))

		sender(txctx)

		errs := txctx.Envelope.GetErrors()
		assert.NotEmpty(t, errs)
	})
}

func TestSender_TesseraTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	envelopeId := utils.RandomString(12)
	txHash := "0x" + utils.RandHexString(64)
	txRaw := "0x" + utils.RandHexString(10)

	schedulerClient := mock2.NewMockTransactionSchedulerClient(ctrl)

	ec := mocks.NewMockEthClient(ctrl)
	sender := Sender(ec, schedulerClient)

	t.Run("should execute Tessera private transaction successfully", func(t *testing.T) {
		nextJobUUID := "nextJobUUID"
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.
			SetJobType(tx.JobType_ETH_TESSERA_PRIVATE_TX).
			SetPrivateFor([]string{"SetPrivateFor=="}).
			SetPrivateFrom("privateFrom==").
			SetNextJobUUID(nextJobUUID).
			SetDataString("0x0123")

		ec.EXPECT().StoreRaw(txctx.Context(), chainRegistryUrl+"/tessera/",
			txctx.Envelope.MustGetDataBytes(), txctx.Envelope.GetPrivateFrom()).
			Return(txHash, nil)

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))
		
		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusStored,
				Transaction: &entities.ETHTransaction{
					EnclaveKey: txHash,
				},
			}))
		
		sender(txctx)
	})

	t.Run("should execute Tessera Marking transaction successfully", func(t *testing.T) {
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.
			SetJobType(tx.JobType_ETH_TESSERA_MARKING_TX).
			SetPrivateFor([]string{"SetPrivateFor=="}).
			SetPrivateFrom("privateFrom==")

		ec.EXPECT().SendQuorumRawPrivateTransaction(txctx.Context(), chainRegistryUrl, txRaw,
			types.Call2PrivateArgs(txctx.Envelope).PrivateFor).
			Return(ethcommon.HexToHash(txHash), nil)

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))

		sender(txctx)
	})

	t.Run("should fail to execute Tessera Marking transaction with missing PrivateFor", func(t *testing.T) {
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.
			SetJobType(tx.JobType_ETH_TESSERA_MARKING_TX).
			SetPrivateFrom("privateFrom==")

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))

		sender(txctx)

		errs := txctx.Envelope.GetErrors()
		assert.NotEmpty(t, errs)
		assert.True(t, errors.IsDataError(errs[0]))
	})
}

func TestSender_EEAPrivateTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// ctx := context.Background()
	envelopeId := utils.RandomString(12)
	txHash := "0x" + utils.RandHexString(64)
	txRaw := "0x" + utils.RandHexString(10)

	schedulerClient := mock2.NewMockTransactionSchedulerClient(ctrl)

	ec := mocks.NewMockEthClient(ctrl)
	sender := Sender(ec, schedulerClient)

	t.Run("should execute eea marking transaction", func(t *testing.T) {
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.SetJobType(tx.JobType_ETH_ORION_MARKING_TX)

		ec.EXPECT().SendRawTransaction(txctx.Context(), chainRegistryUrl, txRaw).Return(ethcommon.HexToHash(txHash), nil)

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))

		sender(txctx)
	})
	t.Run("should execute eea private transaction", func(t *testing.T) {
		nextJobUUID := "nextJobUUID"
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.
			SetJobType(tx.JobType_ETH_ORION_EEA_TX).
			SetPrivateFor([]string{"SetPrivateFor=="}).
			SetNextJobUUID(nextJobUUID).
			SetDataString("0x0123")

		ec.EXPECT().PrivDistributeRawTransaction(txctx.Context(), chainRegistryUrl, txRaw).
			Return(ethcommon.HexToHash(txHash), nil)

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))
		
		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusStored,
				Transaction: &entities.ETHTransaction{
					Hash: txHash,
				},
			}))

		sender(txctx)
	})

	t.Run("should fail execute raw transaction", func(t *testing.T) {
		txctx := newTxCtx(envelopeId, txHash, txRaw)
		_ = txctx.Envelope.
			SetJobType(tx.JobType_ETH_ORION_EEA_TX).
			SetPrivateFor([]string{"SetPrivateFor=="}).
			SetDataString("0x0123")
		err := fmt.Errorf("failed to send a raw transaction")

		ec.EXPECT().PrivDistributeRawTransaction(txctx.Context(), chainRegistryUrl, txRaw).
			Return(ethcommon.Hash{}, err)

		schedulerClient.EXPECT().UpdateJob(txctx.Context(), envelopeId, gomockUpdateStatusMatcher(
			&txschedulertypes.UpdateJobRequest{
				Status: utils.StatusPending,
			}))

		sender(txctx)

		errs := txctx.Envelope.GetErrors()
		assert.NotEmpty(t, errs)
	})
}
