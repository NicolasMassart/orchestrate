package multi

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/src/api/store"
	storePG "github.com/consensys/orchestrate/src/api/store/postgres"
	"github.com/consensys/orchestrate/src/infra/database/postgres"
)

const storeServiceName = "api.store"

func Build(ctx context.Context, cfg *Config, pgmngr postgres.Manager) (store.DB, error) {
	switch cfg.Type {
	case postgresType:
		cfg.Postgres.PG.ApplicationName = storeServiceName
		opts, err := cfg.Postgres.PG.PGOptions()
		if err != nil {
			return nil, err
		}
		return storePG.NewPGDB(pgmngr.Connect(ctx, opts)), nil
	default:
		return nil, fmt.Errorf("invalid API store type %q", cfg.Type)
	}
}
