package migrations

import (
	"github.com/go-pg/migrations/v7"
	log "github.com/sirupsen/logrus"
)

func createEventStreamsTable(db migrations.DB) error {
	log.Debug("Creating event streams table...")

	_, err := db.Exec(`
CREATE TABLE event_streams (
	id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    tenant_id TEXT NOT NULL,
    owner_id TEXT,
    name TEXT,
    chain_uuid UUID,
    channel TEXT NOT NULL,
    status TEXT NOT NULL,
	specs JSONB NOT NULL,
    labels JSON,
	created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc') NOT NULL, 
	updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc') NOT NULL
);

CREATE UNIQUE INDEX event_streams_unique_name_idx ON event_streams (tenant_id, owner_id, name) WHERE name IS NOT NULL;
CREATE UNIQUE INDEX event_streams_unique_chain_idx ON event_streams (tenant_id, owner_id, chain_uuid) WHERE chain_uuid IS NOT NULL;
CREATE UNIQUE INDEX event_streams_unique_no_chain_idx ON event_streams (tenant_id, owner_id) WHERE chain_uuid IS NULL;
`)
	if err != nil {
		log.WithError(err).Error("Could not create event_streams table")
		return err
	}

	log.Info("Created event_streams table")

	return nil
}

func dropEventStreamsTable(db migrations.DB) error {
	log.Debug("Dropping event_streams table")

	_, err := db.Exec(`DROP TABLE event_streams;`)
	if err != nil {
		log.WithError(err).Error("Could not drop event_streams table")
		return err
	}

	log.Info("Dropped event_streams table")

	return nil
}

func init() {
	Collection.MustRegisterTx(createEventStreamsTable, dropEventStreamsTable)
}
