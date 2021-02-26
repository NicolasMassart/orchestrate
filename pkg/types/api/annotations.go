package api

import (
	"github.com/ConsenSys/orchestrate/pkg/utils"
)

type Annotations struct {
	OneTimeKey     bool           `json:"oneTimeKey,omitempty" example:"true"`
	HasBeenRetried bool           `json:"hasBeenRetried,omitempty" example:"false"`
	GasPricePolicy GasPriceParams `json:"gasPricePolicy,omitempty"`
}

func (g *Annotations) Validate() error {
	if err := utils.GetValidator().Struct(g); err != nil {
		return err
	}

	if err := g.GasPricePolicy.RetryPolicy.Validate(); err != nil {
		return err
	}

	return nil
}
