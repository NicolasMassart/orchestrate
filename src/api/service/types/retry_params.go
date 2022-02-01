package types

import (
	"math"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
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
