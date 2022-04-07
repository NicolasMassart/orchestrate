package config

import (
	"github.com/consensys/orchestrate/cmd/flags"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(testDataViperKey, testDataDefault)
	_ = viper.BindEnv(testDataViperKey, testDataEnv)
}

var (
	testDataViperKey = "test.data"
	testDataEnv      = "TEST_GLOBAL_DATA"
	testDataDefault  = "{}"
)

// Flags register Aliases flags
func EnvInit() {
	_ = flags.KafkaFlags
	_ = flags.KafkaConsumerFlags
	_ = orchestrateclient.Flags
	_ = multitenancy.Flags
	_ = log.Flags
}
