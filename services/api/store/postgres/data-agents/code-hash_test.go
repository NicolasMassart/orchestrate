// +build unit
// +build !race
// +build !integration

package dataagents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	pgTestUtils "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/database/postgres/testutils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/store/models"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/store/postgres/migrations"
)

type codeHashTestSuite struct {
	suite.Suite
	agents *PGAgents
	pg     *pgTestUtils.PGTestHelper
}

func TestPGCodeHash(t *testing.T) {
	s := new(codeHashTestSuite)
	suite.Run(t, s)
}

func (s *codeHashTestSuite) SetupSuite() {
	s.pg, _ = pgTestUtils.NewPGTestHelper(nil, migrations.Collection)
	s.pg.InitTestDB(s.T())
}

func (s *codeHashTestSuite) SetupTest() {
	s.pg.UpgradeTestDB(s.T())
	s.agents = New(s.pg.DB)
}

func (s *codeHashTestSuite) TearDownTest() {
	s.pg.DowngradeTestDB(s.T())
}

func (s *codeHashTestSuite) TearDownSuite() {
	s.pg.DropTestDB(s.T())
}

func (s *codeHashTestSuite) TestPGCodeHash_Insert() {
	ctx := context.Background()

	s.T().Run("should insert model successfully", func(t *testing.T) {
		codehash := &models.CodehashModel{
			ChainID:  "chainID",
			Address:  "address",
			Codehash: "codeHash",
		}
		err := s.agents.CodeHash().Insert(ctx, codehash)

		assert.NoError(t, err)
		assert.Equal(t, codehash.ID, 1)
	})

	s.T().Run("should return PostgresConnectionError if insert fails", func(t *testing.T) {
		// We drop the DB to make the test fail
		s.pg.DropTestDB(t)

		codehash := &models.CodehashModel{
			ChainID:  "chainID",
			Address:  "address",
			Codehash: "codeHash",
		}
		err := s.agents.CodeHash().Insert(ctx, codehash)

		assert.True(t, errors.IsInternalError(err))

		// We bring it back up
		s.pg.InitTestDB(t)
	})
}
