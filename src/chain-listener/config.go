package chainlistener

import (
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
)

const (
	DecodedOutTopic = "tx-decoded"
)

type Config struct {
	IsMultiTenancyEnabled bool
	App                   *app.Config
	HTTPClient            *http.Config
	API                   *orchestrateclient.Config
	ChainListenerConfig   *TxListenerConfig
}

type TxListenerConfig struct {
	RefreshInterval time.Duration
	DecodedOutTopic string
}

func NewTxListenerConfig(interval time.Duration, topic string) *TxListenerConfig {
	return &TxListenerConfig{
		RefreshInterval: interval,
		DecodedOutTopic: topic,
	}
}
