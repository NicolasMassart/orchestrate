package cucumber

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/cucumber/godog"
	"github.com/spf13/viper"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/auth/jwt/generator"
	broker "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/broker/sarama"
	ethclient "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/ethclient/rpc"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/log"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/multitenancy"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/sdk/client"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/tests/service/e2e/cucumber/alias"
)

var (
	options  *godog.Options
	initOnce = &sync.Once{}
)

// Init initialize Cucumber service
func Init(ctx context.Context, rawTestData string) {
	initOnce.Do(func() {
		if options != nil {
			return
		}

		logger := log.FromContext(ctx)

		// Initialize Steps
		broker.InitSyncProducer(ctx)
		generator.Init(ctx)
		alias.Init(rawTestData)
		client.Init()
		ethclient.Init(ctx)

		tags := listTagCucumber()

		options = &godog.Options{
			ShowStepDefinitions: viper.GetBool(ShowStepDefinitionsViperKey),
			Randomize:           viper.GetInt64(RandomizeViperKey),
			StopOnFailure:       viper.GetBool(StopOnFailureViperKey),
			Strict:              viper.GetBool(StrictViperKey),
			NoColors:            viper.GetBool(NoColorsViperKey),
			Tags:                tags,
			Format:              viper.GetString(FormatViperKey),
			Concurrency:         viper.GetInt(ConcurrencyViperKey),
			Paths:               viper.GetStringSlice(PathsViperKey),
		}

		outputPath := viper.GetString(OutputPathViperKey)
		if outputPath != "" {
			f, err := os.Create(viper.GetString(OutputPathViperKey))
			if err != nil {
				logger.WithError(err).Fatalf("could not write output in %s", outputPath)
			}
			options.Output = f
		}

		logger.WithField("tags", options.Tags).
			WithField("concurrency", options.Concurrency).
			WithField("paths", options.Paths).
			WithField("output", outputPath).
			Info("service ready")
	})
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

// SetGlobalOptions sets global Cucumber Handler
func SetGlobalOptions(o *godog.Options) {
	options = o
}

// GlobalHandler returns global Cucumber handler
func GlobalOptions() *godog.Options {
	return options
}
