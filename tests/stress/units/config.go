package units

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/tests/stress/assets"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type WorkloadConfig struct {
	accounts               []ethcommon.Address
	chains                 []assets.RegisteredChainData
	artifacts              []string
	privacyGroups          []assets.PrivacyGroup
	waitForEnvelopeTimeout time.Duration
	kafkaTopic             string
}

func NewWorkloadConfig(ctx context.Context, waitForEnvelopeTimeout time.Duration, kafkaTopic string) *WorkloadConfig {
	return &WorkloadConfig{
		accounts:               assets.ContextAccounts(ctx),
		chains:                 assets.ContextChains(ctx),
		artifacts:              assets.ContextArtifacts(ctx),
		privacyGroups:          assets.ContextPrivacyGroups(ctx),
		waitForEnvelopeTimeout: waitForEnvelopeTimeout,
		kafkaTopic:             kafkaTopic,
	}
}
