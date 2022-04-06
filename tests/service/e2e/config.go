package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	broker "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/tests/config"
	"github.com/cucumber/godog"
	"github.com/spf13/viper"
)

type Config struct {
	Cucumber *godog.Options
	Output   string
	TestData *config.TestData
	KafkaCfg *broker.Config
	Timeout  time.Duration
}

func NewConfig(vipr *viper.Viper) (*Config, error) {
	cfg := &Config{}
	rawTestData := viper.GetString(e2eDataViperKey)
	err := json.Unmarshal([]byte(rawTestData), cfg.TestData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test data")
	}

	cfg.KafkaCfg = flags.NewKafkaConfig(vipr)
	cfg.Output = viper.GetString(OutputPathViperKey)
	cfg.Timeout = viper.GetDuration(CucumberTimeoutViperKey)
	cfg.Cucumber = &godog.Options{
		ShowStepDefinitions: viper.GetBool(ShowStepDefinitionsViperKey),
		Randomize:           viper.GetInt64(RandomizeViperKey),
		StopOnFailure:       viper.GetBool(StopOnFailureViperKey),
		Strict:              viper.GetBool(StrictViperKey),
		NoColors:            viper.GetBool(NoColorsViperKey),
		Tags:                listTagCucumber(),
		Format:              viper.GetString(FormatViperKey),
		Concurrency:         viper.GetInt(ConcurrencyViperKey),
		Paths:               viper.GetStringSlice(PathsViperKey),
	}

	cfg.Cucumber.Output, err = os.Create(cfg.Output)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func listTagCucumber() string {
	var tags []string
	if viper.GetString(TagsViperKey) != "" {
		tags = append(tags, strings.Split(viper.GetString(TagsViperKey), " ")...)
	}

	if !viper.GetBool(multitenancy.EnabledViperKey) {
		tags = append(tags, "~@multi-tenancy")
	}

	return strings.Join(tags, " && ")
}
