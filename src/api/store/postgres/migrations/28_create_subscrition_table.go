package migrations

import (
	"github.com/go-pg/migrations/v7"
	log "github.com/sirupsen/logrus"
)

func createSubscriptionsTable(db migrations.DB) error {
	log.Debug("Creating subscriptions table...")

	_, err := db.Exec(`
CREATE TABLE subscriptions (
	id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
	address TEXT,
    tenant_id TEXT NOT NULL,
    owner_id TEXT,
    contract_name TEXT,
    contract_tag TEXT,
    chain_uuid UUID,
    event_stream_uuid UUID NOT NULL,
	created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc') NOT NULL, 
	updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc') NOT NULL
);

CREATE UNIQUE INDEX subscription_unique_address_chain__idx ON subscriptions (address, chain_uuid, tenant_id, owner_id);
`)
	if err != nil {
		log.WithError(err).Error("Could not create subscriptions table")
		return err
	}

	log.Info("Created subscriptions table")

	return nil
}

func dropSubscriptionsTable(db migrations.DB) error {
	log.Debug("Dropping subscriptions table")

	_, err := db.Exec(`DROP TABLE subscriptions;`)
	if err != nil {
		log.WithError(err).Error("Could not drop subscriptions table")
		return err
	}

	log.Info("Dropped subscription table")

	return nil
}

func init() {
	Collection.MustRegisterTx(createSubscriptionsTable, dropSubscriptionsTable)
}
