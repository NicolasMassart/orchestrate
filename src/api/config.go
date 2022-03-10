package api

import (
	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/proxy"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"
	quorumkeymanager "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	"github.com/spf13/viper"
)

type Config struct {
	App          *app.Config
	Postgres     *gopg.Config
	Multitenancy bool
	Proxy        *proxy.Config
	QKM          *quorumkeymanager.Config
}

func NewConfig(vipr *viper.Viper) *Config {
	return &Config{
		App:          app.NewConfig(vipr),
		Postgres:     flags.NewPGConfig(vipr),
		Multitenancy: vipr.GetBool(multitenancy.EnabledViperKey),
		Proxy:        proxy.NewConfig(),
		QKM:          flags.NewQKMConfig(vipr),
	}
}
