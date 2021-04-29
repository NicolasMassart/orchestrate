package healthcheck

import (
	"math"

	"github.com/ConsenSys/orchestrate/pkg/toolkit/app/http"
	"github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/config/dynamic"
	traefikdynamic "github.com/containous/traefik/v2/pkg/config/dynamic"
)

func AddDynamicConfig(cfg *dynamic.Configuration) {
	// Router to Healthchecks
	cfg.HTTP.Routers["healthcheck"] = &dynamic.Router{
		Router: &traefikdynamic.Router{
			EntryPoints: []string{http.DefaultMetricsEntryPoint},
			Service:     "healthcheck",
			Priority:    math.MaxInt32 - 1,
			Rule:        "PathPrefix(`/`)",
		},
	}

	// Healthcheck
	cfg.HTTP.Services["healthcheck"] = &dynamic.Service{
		HealthCheck: &dynamic.HealthCheck{},
	}
}
