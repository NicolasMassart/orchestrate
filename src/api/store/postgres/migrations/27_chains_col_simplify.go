package migrations

import (
	"github.com/go-pg/migrations/v7"
	log "github.com/sirupsen/logrus"
)

func upgradeChainColumnsSimplification(db migrations.DB) error {
	log.Debug("Removing unused chain columns")

	_, err := db.Exec(`
ALTER TABLE chains
	DROP COLUMN listener_current_block,
	DROP COLUMN listener_starting_block,
	DROP COLUMN listener_external_tx_enabled;

ALTER TABLE chains
	RENAME COLUMN listener_back_off_duration TO listener_block_time_duration;
`)
	if err != nil {
		return err
	}

	log.Info("unused chain columns were removed")

	return nil
}

func downgradeChainColumnsSimplification(db migrations.DB) error {
	log.Debug("Undoing removal of unused columns")

	_, err := db.Exec(`
ALTER TABLE chains
	ADD COLUMN listener_current_block BIGINT,
	ADD COLUMN listener_starting_block BIGINT,
	ADD COLUMN listener_external_tx_enabled BOOLEAN DEFAULT false NOT NULL;

ALTER TABLE chains
	RENAME COLUMN listener_block_time_duration TO listener_back_off_duration;
`)
	if err != nil {
		return err
	}

	log.Info("Unused columns were restored")

	return nil
}

func init() {
	Collection.MustRegisterTx(upgradeChainColumnsSimplification, downgradeChainColumnsSimplification)
}
