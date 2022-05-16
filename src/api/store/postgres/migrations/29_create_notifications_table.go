package migrations

import (
	"github.com/go-pg/migrations/v7"
	log "github.com/sirupsen/logrus"
)

func createNotificationsTable(db migrations.DB) error {
	log.Debug("Creating notifications table...")

	_, err := db.Exec(`
CREATE TABLE notifications (
	id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    source_uuid UUID NOT NULL,
    source_type TEXT NOT NULL,
    type TEXT,
    status TEXT NOT NULL,
    api_version TEXT NOT NULL,
	error TEXT,
	created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc') NOT NULL, 
	updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc') NOT NULL,
	UNIQUE(uuid),
	UNIQUE(source_uuid)
);
`)
	if err != nil {
		log.WithError(err).Error("Could not create notifications table")
		return err
	}

	log.Info("Created notifications table")

	return nil
}

func dropNotificationsTable(db migrations.DB) error {
	log.Debug("Dropping notifications table")

	_, err := db.Exec(`DROP TABLE notifications;`)
	if err != nil {
		log.WithError(err).Error("Could not drop notifications table")
		return err
	}

	log.Info("Dropped notifications table")

	return nil
}

func init() {
	Collection.MustRegisterTx(createNotificationsTable, dropNotificationsTable)
}
