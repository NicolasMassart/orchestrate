// Code generated by MockGen. DO NOT EDIT.
// Source: subscriptions.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	entities "github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockCreatedSubscription is a mock of CreatedSubscription interface
type MockCreatedSubscription struct {
	ctrl     *gomock.Controller
	recorder *MockCreatedSubscriptionMockRecorder
}

// MockCreatedSubscriptionMockRecorder is the mock recorder for MockCreatedSubscription
type MockCreatedSubscriptionMockRecorder struct {
	mock *MockCreatedSubscription
}

// NewMockCreatedSubscription creates a new mock instance
func NewMockCreatedSubscription(ctrl *gomock.Controller) *MockCreatedSubscription {
	mock := &MockCreatedSubscription{ctrl: ctrl}
	mock.recorder = &MockCreatedSubscriptionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCreatedSubscription) EXPECT() *MockCreatedSubscriptionMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockCreatedSubscription) Execute(ctx context.Context, sub *entities.Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, sub)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockCreatedSubscriptionMockRecorder) Execute(ctx, sub interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockCreatedSubscription)(nil).Execute), ctx, sub)
}

// MockUpdatedSubscription is a mock of UpdatedSubscription interface
type MockUpdatedSubscription struct {
	ctrl     *gomock.Controller
	recorder *MockUpdatedSubscriptionMockRecorder
}

// MockUpdatedSubscriptionMockRecorder is the mock recorder for MockUpdatedSubscription
type MockUpdatedSubscriptionMockRecorder struct {
	mock *MockUpdatedSubscription
}

// NewMockUpdatedSubscription creates a new mock instance
func NewMockUpdatedSubscription(ctrl *gomock.Controller) *MockUpdatedSubscription {
	mock := &MockUpdatedSubscription{ctrl: ctrl}
	mock.recorder = &MockUpdatedSubscriptionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUpdatedSubscription) EXPECT() *MockUpdatedSubscriptionMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockUpdatedSubscription) Execute(ctx context.Context, sub *entities.Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, sub)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockUpdatedSubscriptionMockRecorder) Execute(ctx, sub interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockUpdatedSubscription)(nil).Execute), ctx, sub)
}

// MockDeletedSubscription is a mock of DeletedSubscription interface
type MockDeletedSubscription struct {
	ctrl     *gomock.Controller
	recorder *MockDeletedSubscriptionMockRecorder
}

// MockDeletedSubscriptionMockRecorder is the mock recorder for MockDeletedSubscription
type MockDeletedSubscriptionMockRecorder struct {
	mock *MockDeletedSubscription
}

// NewMockDeletedSubscription creates a new mock instance
func NewMockDeletedSubscription(ctrl *gomock.Controller) *MockDeletedSubscription {
	mock := &MockDeletedSubscription{ctrl: ctrl}
	mock.recorder = &MockDeletedSubscriptionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeletedSubscription) EXPECT() *MockDeletedSubscriptionMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockDeletedSubscription) Execute(ctx context.Context, subUUID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, subUUID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockDeletedSubscriptionMockRecorder) Execute(ctx, subUUID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockDeletedSubscription)(nil).Execute), ctx, subUUID)
}

// MockNotifySubscriptionEvents is a mock of NotifySubscriptionEvents interface
type MockNotifySubscriptionEvents struct {
	ctrl     *gomock.Controller
	recorder *MockNotifySubscriptionEventsMockRecorder
}

// MockNotifySubscriptionEventsMockRecorder is the mock recorder for MockNotifySubscriptionEvents
type MockNotifySubscriptionEventsMockRecorder struct {
	mock *MockNotifySubscriptionEvents
}

// NewMockNotifySubscriptionEvents creates a new mock instance
func NewMockNotifySubscriptionEvents(ctrl *gomock.Controller) *MockNotifySubscriptionEvents {
	mock := &MockNotifySubscriptionEvents{ctrl: ctrl}
	mock.recorder = &MockNotifySubscriptionEventsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNotifySubscriptionEvents) EXPECT() *MockNotifySubscriptionEventsMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockNotifySubscriptionEvents) Execute(ctx context.Context, chainUUID string, address common.Address, events []types.Log) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, chainUUID, address, events)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockNotifySubscriptionEventsMockRecorder) Execute(ctx, chainUUID, address, events interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockNotifySubscriptionEvents)(nil).Execute), ctx, chainUUID, address, events)
}

// MockSubscriptionUseCases is a mock of SubscriptionUseCases interface
type MockSubscriptionUseCases struct {
	ctrl     *gomock.Controller
	recorder *MockSubscriptionUseCasesMockRecorder
}

// MockSubscriptionUseCasesMockRecorder is the mock recorder for MockSubscriptionUseCases
type MockSubscriptionUseCasesMockRecorder struct {
	mock *MockSubscriptionUseCases
}

// NewMockSubscriptionUseCases creates a new mock instance
func NewMockSubscriptionUseCases(ctrl *gomock.Controller) *MockSubscriptionUseCases {
	mock := &MockSubscriptionUseCases{ctrl: ctrl}
	mock.recorder = &MockSubscriptionUseCasesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSubscriptionUseCases) EXPECT() *MockSubscriptionUseCasesMockRecorder {
	return m.recorder
}

// CreatedSubscriptionUseCase mocks base method
func (m *MockSubscriptionUseCases) CreatedSubscriptionUseCase() usecases.CreatedSubscription {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatedSubscriptionUseCase")
	ret0, _ := ret[0].(usecases.CreatedSubscription)
	return ret0
}

// CreatedSubscriptionUseCase indicates an expected call of CreatedSubscriptionUseCase
func (mr *MockSubscriptionUseCasesMockRecorder) CreatedSubscriptionUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatedSubscriptionUseCase", reflect.TypeOf((*MockSubscriptionUseCases)(nil).CreatedSubscriptionUseCase))
}

// UpdatedSubscriptionUseCase mocks base method
func (m *MockSubscriptionUseCases) UpdatedSubscriptionUseCase() usecases.UpdatedSubscription {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatedSubscriptionUseCase")
	ret0, _ := ret[0].(usecases.UpdatedSubscription)
	return ret0
}

// UpdatedSubscriptionUseCase indicates an expected call of UpdatedSubscriptionUseCase
func (mr *MockSubscriptionUseCasesMockRecorder) UpdatedSubscriptionUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatedSubscriptionUseCase", reflect.TypeOf((*MockSubscriptionUseCases)(nil).UpdatedSubscriptionUseCase))
}

// DeletedSubscriptionJobUseCase mocks base method
func (m *MockSubscriptionUseCases) DeletedSubscriptionJobUseCase() usecases.DeletedSubscription {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletedSubscriptionJobUseCase")
	ret0, _ := ret[0].(usecases.DeletedSubscription)
	return ret0
}

// DeletedSubscriptionJobUseCase indicates an expected call of DeletedSubscriptionJobUseCase
func (mr *MockSubscriptionUseCasesMockRecorder) DeletedSubscriptionJobUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletedSubscriptionJobUseCase", reflect.TypeOf((*MockSubscriptionUseCases)(nil).DeletedSubscriptionJobUseCase))
}

// NotifySubscriptionEventsUseCase mocks base method
func (m *MockSubscriptionUseCases) NotifySubscriptionEventsUseCase() usecases.NotifySubscriptionEvents {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NotifySubscriptionEventsUseCase")
	ret0, _ := ret[0].(usecases.NotifySubscriptionEvents)
	return ret0
}

// NotifySubscriptionEventsUseCase indicates an expected call of NotifySubscriptionEventsUseCase
func (mr *MockSubscriptionUseCasesMockRecorder) NotifySubscriptionEventsUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifySubscriptionEventsUseCase", reflect.TypeOf((*MockSubscriptionUseCases)(nil).NotifySubscriptionEventsUseCase))
}
