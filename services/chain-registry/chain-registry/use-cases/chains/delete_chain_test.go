package chains

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mockstore "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store/mock"
)

func TestDeleteChain_ByUUID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	chainAgent := mockstore.NewMockChainAgent(mockCtrl)

	deleteChainUC := NewDeleteChain(chainAgent)
	chainUUID := uuid.Must(uuid.NewV4()).String()

	chainAgent.EXPECT().DeleteChain(gomock.Any(), gomock.Eq(chainUUID), []string{}).Times(1)

	err := deleteChainUC.Execute(context.Background(), chainUUID, []string{})
	assert.NoError(t, err)
}

func TestDeleteChain_ByUUIDAndTenantID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	chainAgent := mockstore.NewMockChainAgent(mockCtrl)

	deleteChainUC := NewDeleteChain(chainAgent)
	chainUUID := uuid.Must(uuid.NewV4()).String()
	tenantID := "tenantID_1"

	chainAgent.EXPECT().DeleteChain(gomock.Any(), gomock.Eq(chainUUID), []string{tenantID}).Times(1)

	err := deleteChainUC.Execute(context.Background(), chainUUID, []string{tenantID})
	assert.NoError(t, err)
}