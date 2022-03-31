package proto

import (
	"github.com/consensys/orchestrate/pkg/errors"
	ierror "github.com/consensys/orchestrate/pkg/types/error"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
)

func (x *TxResponse) AppendError(err error) *TxResponse {
	x.Errors = append(x.Errors, errors.FromError(err))
	return x
}

func NewTxResponse(job *entities.Job, errs ...*ierror.Error) *TxResponse {
	return &TxResponse{
		Id:            job.ScheduleUUID,
		JobUUID:       job.UUID,
		ContextLabels: job.Labels,
		Transaction: &ethereum.Transaction{
			From:       utils.StringerToString(job.Transaction.From),
			Nonce:      utils.ValueToString(job.Transaction.Nonce),
			To:         utils.StringerToString(job.Transaction.To),
			Value:      utils.StringerToString(job.Transaction.Value),
			Gas:        utils.ValueToString(job.Transaction.Gas),
			GasPrice:   utils.StringerToString(job.Transaction.GasPrice),
			GasFeeCap:  utils.StringerToString(job.Transaction.GasFeeCap),
			GasTipCap:  utils.StringerToString(job.Transaction.GasTipCap),
			Data:       utils.StringerToString(job.Transaction.Data),
			Raw:        utils.StringerToString(job.Transaction.Raw),
			TxHash:     utils.StringerToString(job.Transaction.Hash),
			AccessList: entities.ConvertFromAccessList(job.Transaction.AccessList),
			TxType:     string(job.Transaction.TransactionType),
		},
		Receipt: job.Receipt,
		Chain:   job.ChainUUID,
		Errors:  errs,
	}
}
