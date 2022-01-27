package testdata

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

func FakeTxRequest() *entities.TxRequest {
	return &entities.TxRequest{
		Schedule:       FakeSchedule(),
		IdempotencyKey: "IdempotencyKey",
		ChainName:      "chain",
		Params:         FakeETHTransactionParams(),
		CreatedAt:      time.Now(),
		InternalData:   FakeInternalData(),
	}
}

func FakeRawTxRequest() *entities.TxRequest {
	return &entities.TxRequest{
		Schedule:       FakeSchedule(),
		IdempotencyKey: "IdempotencyKey",
		ChainName:      "chain",
		Params:         FakeRawTransactionParams(),
		CreatedAt:      time.Now(),
		InternalData:   FakeInternalData(),
	}
}

func FakeTransferTxRequest() *entities.TxRequest {
	return &entities.TxRequest{
		Schedule:       FakeSchedule(),
		IdempotencyKey: "IdempotencyKey",
		ChainName:      "chain",
		Params:         FakeTransferTransactionParams(),
		CreatedAt:      time.Now(),
		InternalData:   FakeInternalData(),
	}
}

func FakeTesseraTxRequest() *entities.TxRequest {
	return &entities.TxRequest{
		Schedule:       FakeSchedule(),
		IdempotencyKey: "IdempotencyKey",
		ChainName:      "chain",
		Params:         FakeTesseraTransactionParams(),
		CreatedAt:      time.Now(),
		InternalData:   FakeInternalData(),
	}
}

func FakeEEATxRequest() *entities.TxRequest {
	return &entities.TxRequest{
		Schedule:       FakeSchedule(),
		IdempotencyKey: "IdempotencyKey",
		ChainName:      "chain",
		Params:         FakeEEATransactionParams(),
		CreatedAt:      time.Now(),
		InternalData:   FakeInternalData(),
	}
}
