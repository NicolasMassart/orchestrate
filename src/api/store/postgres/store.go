package postgres

import (
	"github.com/consensys/orchestrate/pkg/errors"
	dataagents "github.com/consensys/orchestrate/src/api/store/postgres/data-agents"
	"github.com/consensys/orchestrate/src/infra/database"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
)

type PGDB struct {
	pg.DB
	*dataagents.PGAgents
}

type PGTX struct {
	pg.Tx
	*dataagents.PGAgents
}

func NewPGDB(db pg.DB) *PGDB {
	return &PGDB{
		DB:       db,
		PGAgents: dataagents.New(db),
	}
}

func (db *PGDB) Begin() (database.Tx, error) {
	db.Transaction()
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, errors.DependencyFailureError("failed to start postgres DB transaction").AppendReason(err.Error())
	}

	return &PGTX{
		Tx:       tx,
		PGAgents: dataagents.New(tx),
	}, nil
}

func (pgTx *PGTX) Begin() (database.Tx, error) {
	return &PGTX{
		Tx:       pgTx.Tx,
		PGAgents: pgTx.PGAgents,
	}, nil
}

func (pgTx *PGTX) Commit() error {
	err := pgTx.Tx.Commit()
	if err != nil {
		return errors.DependencyFailureError("failed to commit postgres DB transaction").AppendReason(err.Error())
	}

	return nil
}

func (pgTx *PGTX) Close() error {
	err := pgTx.Tx.Close()
	if err != nil {
		return errors.DependencyFailureError("failed to close postgres DB transaction").AppendReason(err.Error())
	}

	return nil
}

func (pgTx *PGTX) Rollback() error {
	err := pgTx.Tx.Rollback()
	if err != nil {
		return errors.DependencyFailureError("failed to rollback postgres DB transaction").AppendReason(err.Error())
	}

	return nil
}
