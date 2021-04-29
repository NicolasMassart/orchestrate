package accesslog

import (
	"github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/config/dynamic"
	traefikstatic "github.com/containous/traefik/v2/pkg/config/static"
)

func AddDynamicConfig(cfg *dynamic.Configuration, midName string, staticCfg *traefikstatic.Configuration) {
	// Access Log Middleware
	logFormat := ""
	if staticCfg.Log != nil {
		logFormat = staticCfg.Log.Format
	}

	cfg.HTTP.Middlewares[midName] = &dynamic.Middleware{
		AccessLog: &dynamic.AccessLog{
			Format: logFormat,
		},
	}
}
