package api

import (
	"fmt"
	"math"
	"strings"

	traefikdynamic "github.com/containous/traefik/v2/pkg/config/dynamic"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/configwatcher/provider"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/configwatcher/provider/aggregator"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/configwatcher/provider/static"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/http"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/http/config/dynamic"
)

const (
	InternalProvider = "internal"
)

func NewProvider() provider.Provider {
	prvdr := aggregator.New()
	prvdr.AddProvider(NewInternalProvider())
	return prvdr
}

func NewInternalProvider() provider.Provider {
	return static.New(dynamic.NewMessage(InternalProvider, NewInternalConfig()))
}

func NewInternalConfig() *dynamic.Configuration {
	cfg := dynamic.NewConfig()

	pathPrefix := []string{"/transactions", "/schedules", "/jobs", "/accounts", "/faucets", "/contracts"}
	for idx, path := range pathPrefix {
		pathPrefix[idx] = fmt.Sprintf("PathPrefix(`%s`)", path)
	}

	// Router to API
	cfg.HTTP.Routers["api"] = &dynamic.Router{
		Router: &traefikdynamic.Router{
			EntryPoints: []string{http.DefaultHTTPAppEntryPoint},
			Service:     "api",
			Priority:    math.MaxInt32,
			Rule:        strings.Join(pathPrefix, " || "),
			Middlewares: []string{"base@logger-base", "auth@multitenancy"},
		},
	}

	// API
	cfg.HTTP.Services["api"] = &dynamic.Service{
		API: &dynamic.API{},
	}

	return cfg
}