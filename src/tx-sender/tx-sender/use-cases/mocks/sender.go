// Code generated by MockGen. DO NOT EDIT.
// Source: sender.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	entities "github.com/consensys/orchestrate/src/entities"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockSendETHRawTxUseCase is a mock of SendETHRawTxUseCase interface
type MockSendETHRawTxUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockSendETHRawTxUseCaseMockRecorder
}

// MockSendETHRawTxUseCaseMockRecorder is the mock recorder for MockSendETHRawTxUseCase
type MockSendETHRawTxUseCaseMockRecorder struct {
	mock *MockSendETHRawTxUseCase
}

// NewMockSendETHRawTxUseCase creates a new mock instance
func NewMockSendETHRawTxUseCase(ctrl *gomock.Controller) *MockSendETHRawTxUseCase {
	mock := &MockSendETHRawTxUseCase{ctrl: ctrl}
	mock.recorder = &MockSendETHRawTxUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSendETHRawTxUseCase) EXPECT() *MockSendETHRawTxUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockSendETHRawTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockSendETHRawTxUseCaseMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockSendETHRawTxUseCase)(nil).Execute), ctx, job)
}

// MockSendETHTxUseCase is a mock of SendETHTxUseCase interface
type MockSendETHTxUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockSendETHTxUseCaseMockRecorder
}

// MockSendETHTxUseCaseMockRecorder is the mock recorder for MockSendETHTxUseCase
type MockSendETHTxUseCaseMockRecorder struct {
	mock *MockSendETHTxUseCase
}

// NewMockSendETHTxUseCase creates a new mock instance
func NewMockSendETHTxUseCase(ctrl *gomock.Controller) *MockSendETHTxUseCase {
	mock := &MockSendETHTxUseCase{ctrl: ctrl}
	mock.recorder = &MockSendETHTxUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSendETHTxUseCase) EXPECT() *MockSendETHTxUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockSendETHTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockSendETHTxUseCaseMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockSendETHTxUseCase)(nil).Execute), ctx, job)
}

// MockSendEEAPrivateTxUseCase is a mock of SendEEAPrivateTxUseCase interface
type MockSendEEAPrivateTxUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockSendEEAPrivateTxUseCaseMockRecorder
}

// MockSendEEAPrivateTxUseCaseMockRecorder is the mock recorder for MockSendEEAPrivateTxUseCase
type MockSendEEAPrivateTxUseCaseMockRecorder struct {
	mock *MockSendEEAPrivateTxUseCase
}

// NewMockSendEEAPrivateTxUseCase creates a new mock instance
func NewMockSendEEAPrivateTxUseCase(ctrl *gomock.Controller) *MockSendEEAPrivateTxUseCase {
	mock := &MockSendEEAPrivateTxUseCase{ctrl: ctrl}
	mock.recorder = &MockSendEEAPrivateTxUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSendEEAPrivateTxUseCase) EXPECT() *MockSendEEAPrivateTxUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockSendEEAPrivateTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockSendEEAPrivateTxUseCaseMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockSendEEAPrivateTxUseCase)(nil).Execute), ctx, job)
}

// MockSendGoQuorumPrivateTxUseCase is a mock of SendGoQuorumPrivateTxUseCase interface
type MockSendGoQuorumPrivateTxUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockSendGoQuorumPrivateTxUseCaseMockRecorder
}

// MockSendGoQuorumPrivateTxUseCaseMockRecorder is the mock recorder for MockSendGoQuorumPrivateTxUseCase
type MockSendGoQuorumPrivateTxUseCaseMockRecorder struct {
	mock *MockSendGoQuorumPrivateTxUseCase
}

// NewMockSendGoQuorumPrivateTxUseCase creates a new mock instance
func NewMockSendGoQuorumPrivateTxUseCase(ctrl *gomock.Controller) *MockSendGoQuorumPrivateTxUseCase {
	mock := &MockSendGoQuorumPrivateTxUseCase{ctrl: ctrl}
	mock.recorder = &MockSendGoQuorumPrivateTxUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSendGoQuorumPrivateTxUseCase) EXPECT() *MockSendGoQuorumPrivateTxUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockSendGoQuorumPrivateTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockSendGoQuorumPrivateTxUseCaseMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockSendGoQuorumPrivateTxUseCase)(nil).Execute), ctx, job)
}

// MockSendGoQuorumMarkingTxUseCase is a mock of SendGoQuorumMarkingTxUseCase interface
type MockSendGoQuorumMarkingTxUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockSendGoQuorumMarkingTxUseCaseMockRecorder
}

// MockSendGoQuorumMarkingTxUseCaseMockRecorder is the mock recorder for MockSendGoQuorumMarkingTxUseCase
type MockSendGoQuorumMarkingTxUseCaseMockRecorder struct {
	mock *MockSendGoQuorumMarkingTxUseCase
}

// NewMockSendGoQuorumMarkingTxUseCase creates a new mock instance
func NewMockSendGoQuorumMarkingTxUseCase(ctrl *gomock.Controller) *MockSendGoQuorumMarkingTxUseCase {
	mock := &MockSendGoQuorumMarkingTxUseCase{ctrl: ctrl}
	mock.recorder = &MockSendGoQuorumMarkingTxUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSendGoQuorumMarkingTxUseCase) EXPECT() *MockSendGoQuorumMarkingTxUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockSendGoQuorumMarkingTxUseCase) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockSendGoQuorumMarkingTxUseCaseMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockSendGoQuorumMarkingTxUseCase)(nil).Execute), ctx, job)
}
