package app

import (
	traefikstatic "github.com/containous/traefik/v2/pkg/config/static"
	traefiktypes "github.com/containous/traefik/v2/pkg/types"
	"github.com/spf13/viper"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/configwatcher"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/http"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/log"
	metricsregister "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/metrics/registry"
)

type Config struct {
	HTTP    *HTTP
	Watcher *configwatcher.Config
	Log     *log.Config
	Metrics *metricsregister.Config
}

type HTTP struct {
	EntryPoints  traefikstatic.EntryPoints        `description:"Entry points definition." json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	HostResolver *traefiktypes.HostResolverConfig `description:"Enable CNAME Flattening." json:"hostResolver,omitempty" toml:"hostResolver,omitempty" yaml:"hostResolver,omitempty" label:"allowEmpty" export:"true"`
}

func (c *HTTP) TraefikStatic() *traefikstatic.Configuration {
	return &traefikstatic.Configuration{
		EntryPoints:  c.EntryPoints,
		HostResolver: c.HostResolver,
		API: &traefikstatic.API{
			Dashboard: true,
		},
	}
}

func NewConfig(vipr *viper.Viper) *Config {
	return &Config{
		HTTP: &HTTP{
			EntryPoints: http.NewEPsConfig(vipr),
		},
		Watcher: configwatcher.NewConfig(vipr),
		Log:     log.NewConfig(vipr),
		Metrics: metricsregister.NewConfig(vipr),
	}
}
