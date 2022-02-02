package types

import (
	"math"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type GasPriceParams struct {
	Priority    string      `json:"priority,omitempty" validate:"isPriority" example:"very-high"` // Intended transaction velocity, default is `medium`. Transaction gas is adjusted depending on the velocity to optimize the mining time on user request.
	RetryPolicy RetryParams `json:"retryPolicy,omitempty"`
}
type RetryParams struct {
	Interval  string  `json:"interval,omitempty" validate:"omitempty,minDuration=1s" example:"2m"` // Interval between different transaction retries.
	Increment float64 `json:"increment,omitempty" validate:"omitempty" example:"0.05"`             // Gas increment for each retry. For example, set `0.05` for a 5% increment.
	Limit     float64 `json:"limit,omitempty" validate:"omitempty" example:"0.5"`                  // Maximum gas increment over retries. For example, set `0.5` for a 50% gas increment from source transaction.
}
type IntervalRetryParams struct {
	Interval string `json:"interval,omitempty" validate:"omitempty,isDuration" example:"2m"` // Interval between different transaction retries.
}

const SentryMaxRetries = 10

func (g *RetryParams) Validate() error {
	if err := utils.GetValidator().Struct(g); err != nil {
		return err
	}

	// required_with does not work with floats as the 0 value is valid
	if g.Limit > 0 && g.Increment == 0 {
		return errors.InvalidParameterError("fields 'increment' must be specified when 'limit' is set")
	}
	if g.Increment > 0 && g.Limit == 0 {
		return errors.InvalidParameterError("field 'limit' must be specified when 'increment' is set")
	} else if g.Increment > 0 && math.Ceil(g.Limit/g.Increment) > SentryMaxRetries {
		return errors.InvalidParameterError("Maximum amount of retries %d was exceeded. Reduce 'limit' or increase 'increment` values", SentryMaxRetries)
	}

	return nil
}

type TransactionResponse struct {
	UUID           string                  `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"` // UUID of the transaction.
	IdempotencyKey string                  `json:"idempotencyKey" example:"myIdempotencyKey"`           // Idempotency key of the transaction request.
	ChainName      string                  `json:"chain" example:"myChain"`                             // Chain on which the transaction was created.
	Params         *ETHTransactionResponse `json:"params"`
	Jobs           []*JobResponse          `json:"jobs"`                                            // List of jobs in the transaction.
	CreatedAt      time.Time               `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"` // Date and time at which the transaction was created.
}

func validatePrivateTxParams(protocol entities.PrivateTxManagerType, privateFrom, privacyGroupID string, privateFor []string) error {
	if protocol == "" {
		return errors.InvalidParameterError("field 'protocol' cannot be empty")
	}

	if protocol != entities.TesseraChainType && privateFrom == "" {
		return errors.InvalidParameterError("fields 'privateFrom' cannot be empty")
	}

	if privacyGroupID == "" && len(privateFor) == 0 {
		return errors.InvalidParameterError("fields 'privacyGroupId' and 'privateFor' cannot both be empty")
	}

	if len(privateFor) > 0 && privacyGroupID != "" {
		return errors.InvalidParameterError("fields 'privacyGroupId' and 'privateFor' are mutually exclusive")
	}

	return nil
}

func validateTxFromParams(from *ethcommon.Address, oneTimeKey bool) error {
	if from != nil && oneTimeKey {
		return errors.InvalidParameterError("fields 'from' and 'oneTimeKey' are mutually exclusive")
	}

	if from == nil && !oneTimeKey {
		return errors.InvalidParameterError("field 'from' is required")
	}

	return nil
}
