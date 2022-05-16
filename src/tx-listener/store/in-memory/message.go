package inmemory

import (
	"context"
	"sync"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/tx-listener/store"
)

type messengerState struct {
	indexedByJobUUID map[string]*entities.Message
	markedJobList    *utils.SortedList
	mux              *sync.RWMutex
}

func NewMessengerState() store.Message {
	return &messengerState{
		indexedByJobUUID: make(map[string]*entities.Message),
		markedJobList:    utils.NewSortedList(),
		mux:              &sync.RWMutex{},
	}
}

func (m *messengerState) AddJobMessage(_ context.Context, jobUUID string, msg *entities.Message) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.indexedByJobUUID[jobUUID]; ok {
		return errors.AlreadyExistsError("message %q is duplicated", jobUUID)
	}

	m.indexedByJobUUID[jobUUID] = msg
	return nil
}

func (m *messengerState) GetJobMessage(_ context.Context, jobUUID string) (*entities.Message, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	if msg, ok := m.indexedByJobUUID[jobUUID]; ok {
		return msg, nil
	}

	return nil, errors.NotFoundError("message %q is not found", jobUUID)
}

func (m *messengerState) MarkJobMessage(_ context.Context, jobUUID string) (int, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	if _, ok := m.indexedByJobUUID[jobUUID]; !ok {
		return -1, errors.NotFoundError("message %q is not found", jobUUID)
	}

	msg := m.indexedByJobUUID[jobUUID]
	idx := m.markedJobList.Append(jobUUID, msg.Offset)
	return idx, nil
}

func (m *messengerState) GetMarkedJobMessageByOffset(_ context.Context, offset int64) (string, *entities.Message, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	jobUUID, err := m.markedJobList.GetByKey(offset)
	if err != nil {
		return "", nil, err
	}
	if jobUUID == nil {
		return "", nil, nil
	}

	msg := m.indexedByJobUUID[jobUUID.(string)]
	return jobUUID.(string), msg, nil
}

func (m *messengerState) RemoveJobMessage(_ context.Context, jobUUID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if msg, ok := m.indexedByJobUUID[jobUUID]; ok {
		_ = m.markedJobList.RemoveByKey(msg.Offset)
		delete(m.indexedByJobUUID, jobUUID)
		return nil
	}

	return errors.NotFoundError("message %q is not found", jobUUID)
}
