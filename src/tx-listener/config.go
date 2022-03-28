package txlistener

import (
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
)

type Config struct {
	IsMultiTenancyEnabled bool
	App                   *app.Config
	HTTPClient            *http.Config
	API                   *orchestrateclient.Config
	RefreshInterval       time.Duration
}
