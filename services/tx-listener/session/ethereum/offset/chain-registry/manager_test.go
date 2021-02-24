// +build unit

package chainregistry

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/client/mock"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store/models"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/tx-listener/dynamic"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/tx-listener/session/ethereum/offset"
)

var mockChain = &models.Chain{
	UUID:                    "test-chain",
	Name:                    "test",
	TenantID:                "test",
	URLs:                    []string{"test"},
	ListenerDepth:           &(&struct{ x uint64 }{0}).x,
	ListenerCurrentBlock:    &(&struct{ x uint64 }{0}).x,
	ListenerStartingBlock:   &(&struct{ x uint64 }{0}).x,
	ListenerBackOffDuration: &(&struct{ x string }{"0s"}).x,
}

type ManagerTestSuite struct {
	suite.Suite
	Manager offset.Manager
}

var mockChainRegistryClient *mock.MockChainRegistryClient

func (s *ManagerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	mockChainRegistryClient = mock.NewMockChainRegistryClient(ctrl)

	s.Manager = NewManager(mockChainRegistryClient)
}

func (s *ManagerTestSuite) TestManagerLastBlock() {
	updatedCurrentBlock := uint64(1)
	chain := &dynamic.Chain{
		UUID: mockChain.UUID,
		Listener: &dynamic.Listener{
			CurrentBlock: 0,
		},
	}

	mockChainRegistryClient.EXPECT().GetChainByUUID(gomock.Any(), chain.UUID).Return(mockChain, nil)
	mockChainRegistryClient.EXPECT().UpdateBlockPosition(gomock.Any(), chain.UUID, updatedCurrentBlock)

	lastBlockNumber, err := s.Manager.GetLastBlockNumber(context.Background(), chain)
	assert.NoError(s.T(), err, "GetLastBlockNumber should not error")
	assert.Equal(s.T(), *mockChain.ListenerCurrentBlock, lastBlockNumber, "Lastblock should be correct")

	err = s.Manager.SetLastBlockNumber(context.Background(), chain, updatedCurrentBlock)
	assert.NoError(s.T(), err, "SetLastBlockNumber should not error")
}

func (s *ManagerTestSuite) TestManagerLastBlock_ignored() {
	updatedCurrentBlock := uint64(0)
	chain := &dynamic.Chain{
		UUID: mockChain.UUID,
		Listener: &dynamic.Listener{
			CurrentBlock: 0,
		},
	}

	mockChainRegistryClient.EXPECT().GetChainByUUID(gomock.Any(), chain.UUID).Return(mockChain, nil)

	lastBlockNumber, err := s.Manager.GetLastBlockNumber(context.Background(), chain)
	assert.NoError(s.T(), err, "GetLastBlockNumber should not error")
	assert.Equal(s.T(), *mockChain.ListenerCurrentBlock, lastBlockNumber, "Lastblock should be correct")

	err = s.Manager.SetLastBlockNumber(context.Background(), chain, updatedCurrentBlock)
}

func (s *ManagerTestSuite) TestManagerLastIndex() {
	chain := &dynamic.Chain{
		UUID: mockChain.UUID,
	}

	lastTxIndex, err := s.Manager.GetLastTxIndex(context.Background(), chain, 10)
	assert.NoError(s.T(), err, "GetLastTxIndex should not error")
	assert.Equal(s.T(), uint64(0), lastTxIndex, "LastTxIndex should be correct")

	err = s.Manager.SetLastTxIndex(context.Background(), chain, 10, 17)
	assert.NoError(s.T(), err, "SetLastTxIndex should not error")

	lastTxIndex, err = s.Manager.GetLastTxIndex(context.Background(), chain, 10)
	assert.NoError(s.T(), err, "GetLastTxIndex should not error")
	assert.Equal(s.T(), uint64(17), lastTxIndex, "LastTxIndex should be correct")

	lastTxIndex, err = s.Manager.GetLastTxIndex(context.Background(), chain, 11)
	assert.NoError(s.T(), err, "GetLastTxIndex should not error")
	assert.Equal(s.T(), uint64(0), lastTxIndex, "LastTxIndex should be correct")
}

func TestRegistry(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}
