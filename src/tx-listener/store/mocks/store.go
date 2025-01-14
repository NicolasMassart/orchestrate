// Code generated by MockGen. DO NOT EDIT.
// Source: store.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	entities "github.com/consensys/orchestrate/src/entities"
	store "github.com/consensys/orchestrate/src/tx-listener/store"
	common "github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockChain is a mock of Chain interface
type MockChain struct {
	ctrl     *gomock.Controller
	recorder *MockChainMockRecorder
}

// MockChainMockRecorder is the mock recorder for MockChain
type MockChainMockRecorder struct {
	mock *MockChain
}

// NewMockChain creates a new mock instance
func NewMockChain(ctrl *gomock.Controller) *MockChain {
	mock := &MockChain{ctrl: ctrl}
	mock.recorder = &MockChainMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChain) EXPECT() *MockChainMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockChain) Add(ctx context.Context, chain *entities.Chain) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, chain)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add
func (mr *MockChainMockRecorder) Add(ctx, chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockChain)(nil).Add), ctx, chain)
}

// Update mocks base method
func (m *MockChain) Update(ctx context.Context, chain *entities.Chain) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, chain)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockChainMockRecorder) Update(ctx, chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockChain)(nil).Update), ctx, chain)
}

// Delete mocks base method
func (m *MockChain) Delete(ctx context.Context, chainUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, chainUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockChainMockRecorder) Delete(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockChain)(nil).Delete), ctx, chainUUID)
}

// Get mocks base method
func (m *MockChain) Get(ctx context.Context, chainUUID string) (*entities.Chain, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, chainUUID)
	ret0, _ := ret[0].(*entities.Chain)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockChainMockRecorder) Get(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockChain)(nil).Get), ctx, chainUUID)
}

// MockPendingJob is a mock of PendingJob interface
type MockPendingJob struct {
	ctrl     *gomock.Controller
	recorder *MockPendingJobMockRecorder
}

// MockPendingJobMockRecorder is the mock recorder for MockPendingJob
type MockPendingJobMockRecorder struct {
	mock *MockPendingJob
}

// NewMockPendingJob creates a new mock instance
func NewMockPendingJob(ctrl *gomock.Controller) *MockPendingJob {
	mock := &MockPendingJob{ctrl: ctrl}
	mock.recorder = &MockPendingJobMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPendingJob) EXPECT() *MockPendingJobMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockPendingJob) Add(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add
func (mr *MockPendingJobMockRecorder) Add(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockPendingJob)(nil).Add), ctx, job)
}

// Remove mocks base method
func (m *MockPendingJob) Remove(ctx context.Context, jobUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", ctx, jobUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove
func (mr *MockPendingJobMockRecorder) Remove(ctx, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockPendingJob)(nil).Remove), ctx, jobUUID)
}

// Update mocks base method
func (m *MockPendingJob) Update(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockPendingJobMockRecorder) Update(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockPendingJob)(nil).Update), ctx, job)
}

// GetChildrenJobUUIDs mocks base method
func (m *MockPendingJob) GetChildrenJobUUIDs(ctx context.Context, jobUUID string) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChildrenJobUUIDs", ctx, jobUUID)
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetChildrenJobUUIDs indicates an expected call of GetChildrenJobUUIDs
func (mr *MockPendingJobMockRecorder) GetChildrenJobUUIDs(ctx, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChildrenJobUUIDs", reflect.TypeOf((*MockPendingJob)(nil).GetChildrenJobUUIDs), ctx, jobUUID)
}

// GetByTxHash mocks base method
func (m *MockPendingJob) GetByTxHash(ctx context.Context, chainUUID string, txHash *common.Hash) (*entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByTxHash", ctx, chainUUID, txHash)
	ret0, _ := ret[0].(*entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByTxHash indicates an expected call of GetByTxHash
func (mr *MockPendingJobMockRecorder) GetByTxHash(ctx, chainUUID, txHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByTxHash", reflect.TypeOf((*MockPendingJob)(nil).GetByTxHash), ctx, chainUUID, txHash)
}

// GetJobUUID mocks base method
func (m *MockPendingJob) GetJobUUID(ctx context.Context, jobUUID string) (*entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJobUUID", ctx, jobUUID)
	ret0, _ := ret[0].(*entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJobUUID indicates an expected call of GetJobUUID
func (mr *MockPendingJobMockRecorder) GetJobUUID(ctx, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJobUUID", reflect.TypeOf((*MockPendingJob)(nil).GetJobUUID), ctx, jobUUID)
}

// ListPerChainUUID mocks base method
func (m *MockPendingJob) ListPerChainUUID(ctx context.Context, chainUUID string) ([]*entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPerChainUUID", ctx, chainUUID)
	ret0, _ := ret[0].([]*entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPerChainUUID indicates an expected call of ListPerChainUUID
func (mr *MockPendingJobMockRecorder) ListPerChainUUID(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPerChainUUID", reflect.TypeOf((*MockPendingJob)(nil).ListPerChainUUID), ctx, chainUUID)
}

// DeletePerChainUUID mocks base method
func (m *MockPendingJob) DeletePerChainUUID(ctx context.Context, chainUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePerChainUUID", ctx, chainUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePerChainUUID indicates an expected call of DeletePerChainUUID
func (mr *MockPendingJobMockRecorder) DeletePerChainUUID(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePerChainUUID", reflect.TypeOf((*MockPendingJob)(nil).DeletePerChainUUID), ctx, chainUUID)
}

// MockMessage is a mock of Message interface
type MockMessage struct {
	ctrl     *gomock.Controller
	recorder *MockMessageMockRecorder
}

// MockMessageMockRecorder is the mock recorder for MockMessage
type MockMessageMockRecorder struct {
	mock *MockMessage
}

// NewMockMessage creates a new mock instance
func NewMockMessage(ctrl *gomock.Controller) *MockMessage {
	mock := &MockMessage{ctrl: ctrl}
	mock.recorder = &MockMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMessage) EXPECT() *MockMessageMockRecorder {
	return m.recorder
}

// AddJobMessage mocks base method
func (m *MockMessage) AddJobMessage(ctx context.Context, jobUUID string, msg *entities.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddJobMessage", ctx, jobUUID, msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddJobMessage indicates an expected call of AddJobMessage
func (mr *MockMessageMockRecorder) AddJobMessage(ctx, jobUUID, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddJobMessage", reflect.TypeOf((*MockMessage)(nil).AddJobMessage), ctx, jobUUID, msg)
}

// GetJobMessage mocks base method
func (m *MockMessage) GetJobMessage(ctx context.Context, jobUUID string) (*entities.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJobMessage", ctx, jobUUID)
	ret0, _ := ret[0].(*entities.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJobMessage indicates an expected call of GetJobMessage
func (mr *MockMessageMockRecorder) GetJobMessage(ctx, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJobMessage", reflect.TypeOf((*MockMessage)(nil).GetJobMessage), ctx, jobUUID)
}

// MarkJobMessage mocks base method
func (m *MockMessage) MarkJobMessage(arg0 context.Context, jobUUID string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkJobMessage", arg0, jobUUID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarkJobMessage indicates an expected call of MarkJobMessage
func (mr *MockMessageMockRecorder) MarkJobMessage(arg0, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkJobMessage", reflect.TypeOf((*MockMessage)(nil).MarkJobMessage), arg0, jobUUID)
}

// GetMarkedJobMessageByOffset mocks base method
func (m *MockMessage) GetMarkedJobMessageByOffset(arg0 context.Context, offset int64) (string, *entities.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMarkedJobMessageByOffset", arg0, offset)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(*entities.Message)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetMarkedJobMessageByOffset indicates an expected call of GetMarkedJobMessageByOffset
func (mr *MockMessageMockRecorder) GetMarkedJobMessageByOffset(arg0, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMarkedJobMessageByOffset", reflect.TypeOf((*MockMessage)(nil).GetMarkedJobMessageByOffset), arg0, offset)
}

// RemoveJobMessage mocks base method
func (m *MockMessage) RemoveJobMessage(arg0 context.Context, jobUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveJobMessage", arg0, jobUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveJobMessage indicates an expected call of RemoveJobMessage
func (mr *MockMessageMockRecorder) RemoveJobMessage(arg0, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveJobMessage", reflect.TypeOf((*MockMessage)(nil).RemoveJobMessage), arg0, jobUUID)
}

// MockSubscriptions is a mock of Subscriptions interface
type MockSubscriptions struct {
	ctrl     *gomock.Controller
	recorder *MockSubscriptionsMockRecorder
}

// MockSubscriptionsMockRecorder is the mock recorder for MockSubscriptions
type MockSubscriptionsMockRecorder struct {
	mock *MockSubscriptions
}

// NewMockSubscriptions creates a new mock instance
func NewMockSubscriptions(ctrl *gomock.Controller) *MockSubscriptions {
	mock := &MockSubscriptions{ctrl: ctrl}
	mock.recorder = &MockSubscriptionsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSubscriptions) EXPECT() *MockSubscriptionsMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockSubscriptions) Add(ctx context.Context, sub *entities.Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, sub)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add
func (mr *MockSubscriptionsMockRecorder) Add(ctx, sub interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockSubscriptions)(nil).Add), ctx, sub)
}

// Remove mocks base method
func (m *MockSubscriptions) Remove(ctx context.Context, subUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", ctx, subUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove
func (mr *MockSubscriptionsMockRecorder) Remove(ctx, subUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockSubscriptions)(nil).Remove), ctx, subUUID)
}

// Update mocks base method
func (m *MockSubscriptions) Update(ctx context.Context, sub *entities.Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, sub)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockSubscriptionsMockRecorder) Update(ctx, sub interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSubscriptions)(nil).Update), ctx, sub)
}

// ListPerChainUUID mocks base method
func (m *MockSubscriptions) ListPerChainUUID(ctx context.Context, chainUUID string) ([]*entities.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPerChainUUID", ctx, chainUUID)
	ret0, _ := ret[0].([]*entities.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPerChainUUID indicates an expected call of ListPerChainUUID
func (mr *MockSubscriptionsMockRecorder) ListPerChainUUID(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPerChainUUID", reflect.TypeOf((*MockSubscriptions)(nil).ListPerChainUUID), ctx, chainUUID)
}

// ListAddressesPerChainUUID mocks base method
func (m *MockSubscriptions) ListAddressesPerChainUUID(ctx context.Context, chainUUID string) ([]common.Address, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAddressesPerChainUUID", ctx, chainUUID)
	ret0, _ := ret[0].([]common.Address)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAddressesPerChainUUID indicates an expected call of ListAddressesPerChainUUID
func (mr *MockSubscriptionsMockRecorder) ListAddressesPerChainUUID(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAddressesPerChainUUID", reflect.TypeOf((*MockSubscriptions)(nil).ListAddressesPerChainUUID), ctx, chainUUID)
}

// MockRetryJobSession is a mock of RetryJobSession interface
type MockRetryJobSession struct {
	ctrl     *gomock.Controller
	recorder *MockRetryJobSessionMockRecorder
}

// MockRetryJobSessionMockRecorder is the mock recorder for MockRetryJobSession
type MockRetryJobSessionMockRecorder struct {
	mock *MockRetryJobSession
}

// NewMockRetryJobSession creates a new mock instance
func NewMockRetryJobSession(ctrl *gomock.Controller) *MockRetryJobSession {
	mock := &MockRetryJobSession{ctrl: ctrl}
	mock.recorder = &MockRetryJobSessionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRetryJobSession) EXPECT() *MockRetryJobSessionMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockRetryJobSession) Add(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add
func (mr *MockRetryJobSessionMockRecorder) Add(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockRetryJobSession)(nil).Add), ctx, job)
}

// Has mocks base method
func (m *MockRetryJobSession) Has(ctx context.Context, jobUUID string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Has", ctx, jobUUID)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Has indicates an expected call of Has
func (mr *MockRetryJobSessionMockRecorder) Has(ctx, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Has", reflect.TypeOf((*MockRetryJobSession)(nil).Has), ctx, jobUUID)
}

// Remove mocks base method
func (m *MockRetryJobSession) Remove(ctx context.Context, jobUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", ctx, jobUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove
func (mr *MockRetryJobSessionMockRecorder) Remove(ctx, jobUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockRetryJobSession)(nil).Remove), ctx, jobUUID)
}

// GetByTxHash mocks base method
func (m *MockRetryJobSession) GetByTxHash(ctx context.Context, chainUUID string, txHash *common.Hash) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByTxHash", ctx, chainUUID, txHash)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByTxHash indicates an expected call of GetByTxHash
func (mr *MockRetryJobSessionMockRecorder) GetByTxHash(ctx, chainUUID, txHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByTxHash", reflect.TypeOf((*MockRetryJobSession)(nil).GetByTxHash), ctx, chainUUID, txHash)
}

// ListByChainUUID mocks base method
func (m *MockRetryJobSession) ListByChainUUID(ctx context.Context, chainUUID string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByChainUUID", ctx, chainUUID)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByChainUUID indicates an expected call of ListByChainUUID
func (mr *MockRetryJobSessionMockRecorder) ListByChainUUID(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByChainUUID", reflect.TypeOf((*MockRetryJobSession)(nil).ListByChainUUID), ctx, chainUUID)
}

// DeletePerChainUUID mocks base method
func (m *MockRetryJobSession) DeletePerChainUUID(ctx context.Context, chainUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePerChainUUID", ctx, chainUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePerChainUUID indicates an expected call of DeletePerChainUUID
func (mr *MockRetryJobSessionMockRecorder) DeletePerChainUUID(ctx, chainUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePerChainUUID", reflect.TypeOf((*MockRetryJobSession)(nil).DeletePerChainUUID), ctx, chainUUID)
}

// MockState is a mock of State interface
type MockState struct {
	ctrl     *gomock.Controller
	recorder *MockStateMockRecorder
}

// MockStateMockRecorder is the mock recorder for MockState
type MockStateMockRecorder struct {
	mock *MockState
}

// NewMockState creates a new mock instance
func NewMockState(ctrl *gomock.Controller) *MockState {
	mock := &MockState{ctrl: ctrl}
	mock.recorder = &MockStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockState) EXPECT() *MockStateMockRecorder {
	return m.recorder
}

// ChainState mocks base method
func (m *MockState) ChainState() store.Chain {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChainState")
	ret0, _ := ret[0].(store.Chain)
	return ret0
}

// ChainState indicates an expected call of ChainState
func (mr *MockStateMockRecorder) ChainState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChainState", reflect.TypeOf((*MockState)(nil).ChainState))
}

// PendingJobState mocks base method
func (m *MockState) PendingJobState() store.PendingJob {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PendingJobState")
	ret0, _ := ret[0].(store.PendingJob)
	return ret0
}

// PendingJobState indicates an expected call of PendingJobState
func (mr *MockStateMockRecorder) PendingJobState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PendingJobState", reflect.TypeOf((*MockState)(nil).PendingJobState))
}

// RetryJobSessionState mocks base method
func (m *MockState) RetryJobSessionState() store.RetryJobSession {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetryJobSessionState")
	ret0, _ := ret[0].(store.RetryJobSession)
	return ret0
}

// RetryJobSessionState indicates an expected call of RetryJobSessionState
func (mr *MockStateMockRecorder) RetryJobSessionState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetryJobSessionState", reflect.TypeOf((*MockState)(nil).RetryJobSessionState))
}

// SubscriptionState mocks base method
func (m *MockState) SubscriptionState() store.Subscriptions {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscriptionState")
	ret0, _ := ret[0].(store.Subscriptions)
	return ret0
}

// SubscriptionState indicates an expected call of SubscriptionState
func (mr *MockStateMockRecorder) SubscriptionState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscriptionState", reflect.TypeOf((*MockState)(nil).SubscriptionState))
}

// MessengerState mocks base method
func (m *MockState) MessengerState() store.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MessengerState")
	ret0, _ := ret[0].(store.Message)
	return ret0
}

// MessengerState indicates an expected call of MessengerState
func (mr *MockStateMockRecorder) MessengerState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MessengerState", reflect.TypeOf((*MockState)(nil).MessengerState))
}
