package alias

import (
	"encoding/json"
	"sync"

	"github.com/consensys/orchestrate/cmd/flags"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const GlobalAka = "global"
const ExternalTxLabel = "externalTx"

var (
	aliases  *Registry
	initOnce = &sync.Once{}
)

func Init(rawTestData string) {
	initOnce.Do(func() {
		if aliases != nil {
			return
		}

		aliases = NewAliasRegistry()

		// Register global aliases
		importGlobalAlias(rawTestData)
	})
}

// GlobalAliasRegistry returns global Alias registry
func GlobalAliasRegistry() *Registry {
	return aliases
}

func importGlobalAlias(rawTestData string) {
	// register internal aliases
	internal := map[string]interface{}{
		"api":                 viper.GetString(client.URLViperKey),
		"api-metrics":         viper.GetString(client.MetricsURLViperKey),
		"api-key":             viper.GetString(key.APIKeyViperKey),
		"tx-sender-metrics":   viper.GetString(flags.TxSenderMetricsURLViperKey),
		"key-manager":         viper.GetString(flags.URLViperKey),
		"key-manager-metrics": viper.GetString(flags.QKMMetricsURLViperKey),
		"external-tx-label":   ExternalTxLabel,
	}

	// import aliases from environment variable
	global := make(map[string]interface{})
	err := json.Unmarshal([]byte(rawTestData), &global)
	if err != nil {
		log.WithError(err).Fatalf("could not parse and register global")
	}
	for k, v := range internal {
		if _, ok := global[k]; ok {
			log.Fatalf("the key '%s' is not allowed in global alias", k)
		}
		global[k] = v
	}

	aliases.Set(global, GlobalAka)
}
