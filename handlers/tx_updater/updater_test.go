// +build unit

package txupdater

import (
	"fmt"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/engine"
	txschedulertypes "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/txscheduler"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/utils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/client/mock"
	"testing"
)

func TestTransactionUpdater(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	schedulerClient := mock.NewMockTransactionSchedulerClient(mockCtrl)

	t.Run("should do nothing if the tx does not contain errors", func(t *testing.T) {
		txctx := engine.NewTxContext()
		_ = txctx.Envelope.SetID("test")
		txctx.Logger = log.NewEntry(log.New())

		h := TransactionUpdater(schedulerClient)
		h(txctx)
	})
	
	t.Run("should update the status successfully to RECOVERING if envelope contains invalid nonce errors", func(t *testing.T) {
		txctx := engine.NewTxContext()
		_ = txctx.Envelope.SetID("test")
		_ = txctx.AbortWithError(fmt.Errorf("error"))
		txctx.SetInvalidNonceErr(true)
		txctx.Logger = log.NewEntry(log.New())

		schedulerClient.EXPECT().
			UpdateJob(txctx.Context(), txctx.Envelope.GetJobUUID(), &txschedulertypes.UpdateJobRequest{
				Status:  utils.StatusRecovering,
				Message: txctx.Envelope.Error(),
			}).
			Return(&txschedulertypes.JobResponse{}, nil)

		h := TransactionUpdater(schedulerClient)
		h(txctx)
	})

	t.Run("should update the status successfully to FAILED if envelope contains errors", func(t *testing.T) {
		txctx := engine.NewTxContext()
		_ = txctx.Envelope.SetID("test")
		_ = txctx.AbortWithError(fmt.Errorf("error"))
		txctx.Logger = log.NewEntry(log.New())

		schedulerClient.EXPECT().
			UpdateJob(txctx.Context(), txctx.Envelope.GetJobUUID(), &txschedulertypes.UpdateJobRequest{
				Status:  utils.StatusFailed,
				Message: txctx.Envelope.Error(),
			}).
			Return(&txschedulertypes.JobResponse{}, nil)

		h := TransactionUpdater(schedulerClient)
		h(txctx)
	})
}
