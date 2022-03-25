package migrations

import (
	"github.com/go-pg/migrations/v7"
	log "github.com/sirupsen/logrus"
)

func upgradeBytecodeUnique(db migrations.DB) error {
	log.Debug("Removing unique constraint on bytecode")

	_, err := db.Exec(`DROP INDEX unique_abi_bytecode;`)
	if err != nil {
		return err
	}

	log.Info("unique constraint on bytecode removed")

	return nil
}

func downgradeBytecodeUnique(db migrations.DB) error {
	log.Debug("Undoing removal of unique constraint on bytecode")

	_, err := db.Exec(`CREATE UNIQUE INDEX unique_abi_bytecode ON artifacts (md5(abi), md5(deployed_bytecode));`)
	if err != nil {
		return err
	}

	log.Info("unique constraint on bytecode added")

	return nil
}

func init() {
	Collection.MustRegisterTx(upgradeBytecodeUnique, downgradeBytecodeUnique)
}
