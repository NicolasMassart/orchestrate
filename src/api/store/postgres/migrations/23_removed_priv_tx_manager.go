package migrations

import (
	"github.com/go-pg/migrations/v7"
	log "github.com/sirupsen/logrus"
)

func upgradeRemovalPrivateTxMngr(db migrations.DB) error {
	log.Debug("Dropping private_tx_managers tables")

	_, err := db.Exec(`
ALTER TYPE job_type RENAME VALUE 'eth://tessera/privateTransaction' TO 'eth://go-quorum/privateTransaction';
ALTER TYPE job_type RENAME VALUE 'eth://tessera/markingTransaction' TO 'eth://go-quorum/markingTransaction';

ALTER TYPE priv_chain_type RENAME VALUE 'Tessera' TO 'GoQuorum';

ALTER TABLE chains
	ADD COLUMN private_tx_manager_url TEXT;

UPDATE chains
	SET private_tx_manager_url = ptm.url
FROM
  private_tx_managers ptm
WHERE
  chains.uuid=ptm.chain_uuid;

DROP TABLE private_tx_managers;
`)
	if err != nil {
		log.WithError(err).Error("Could not drop private_tx_managers table")
		return err
	}
	log.Info("Dropped private_tx_managers table")

	return nil
}

func downgradeRemovalPrivateTxMngr(db migrations.DB) error {
	log.Debug("Creating private_tx_managers tables...")
	_, err := db.Exec(`

ALTER TYPE job_type RENAME VALUE 'eth://go-quorum/privateTransaction' TO 'eth://tessera/privateTransaction';
ALTER TYPE job_type RENAME VALUE 'eth://go-quorum/markingTransaction' TO 'eth://tessera/markingTransaction';

ALTER TYPE priv_chain_type RENAME VALUE 'GoQuorum' TO 'Tessera';

CREATE TABLE private_tx_managers (
	uuid UUID PRIMARY KEY,
	chain_uuid UUID NOT NULL REFERENCES chains(uuid) ON DELETE CASCADE,
	url TEXT NOT NULL,
	type priv_chain_type NOT NULL,
	created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc') NOT NULL
);
`)
	if err != nil {
		log.WithError(err).Error("Could not create private_tx_managers table")
		return err
	}
	log.Info("Created private_tx_managers table")

	return nil
}

func init() {
	Collection.MustRegisterTx(upgradeRemovalPrivateTxMngr, downgradeRemovalPrivateTxMngr)
}
