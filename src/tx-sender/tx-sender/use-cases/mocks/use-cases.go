// Code generated by MockGen. DO NOT EDIT.
// Source: use-cases.go

// Package mocks is a generated GoMock package.
package mocks

import (
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockUseCases is a mock of UseCases interface
type MockUseCases struct {
	ctrl     *gomock.Controller
	recorder *MockUseCasesMockRecorder
}

// MockUseCasesMockRecorder is the mock recorder for MockUseCases
type MockUseCasesMockRecorder struct {
	mock *MockUseCases
}

// NewMockUseCases creates a new mock instance
func NewMockUseCases(ctrl *gomock.Controller) *MockUseCases {
	mock := &MockUseCases{ctrl: ctrl}
	mock.recorder = &MockUseCasesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUseCases) EXPECT() *MockUseCasesMockRecorder {
	return m.recorder
}

// SendETHRawTx mocks base method
func (m *MockUseCases) SendETHRawTx() usecases.SendETHRawTxUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendETHRawTx")
	ret0, _ := ret[0].(usecases.SendETHRawTxUseCase)
	return ret0
}

// SendETHRawTx indicates an expected call of SendETHRawTx
func (mr *MockUseCasesMockRecorder) SendETHRawTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendETHRawTx", reflect.TypeOf((*MockUseCases)(nil).SendETHRawTx))
}

// SendETHTx mocks base method
func (m *MockUseCases) SendETHTx() usecases.SendETHTxUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendETHTx")
	ret0, _ := ret[0].(usecases.SendETHTxUseCase)
	return ret0
}

// SendETHTx indicates an expected call of SendETHTx
func (mr *MockUseCasesMockRecorder) SendETHTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendETHTx", reflect.TypeOf((*MockUseCases)(nil).SendETHTx))
}

// SendEEAPrivateTx mocks base method
func (m *MockUseCases) SendEEAPrivateTx() usecases.SendEEAPrivateTxUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendEEAPrivateTx")
	ret0, _ := ret[0].(usecases.SendEEAPrivateTxUseCase)
	return ret0
}

// SendEEAPrivateTx indicates an expected call of SendEEAPrivateTx
func (mr *MockUseCasesMockRecorder) SendEEAPrivateTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendEEAPrivateTx", reflect.TypeOf((*MockUseCases)(nil).SendEEAPrivateTx))
}

// SendGoQuorumPrivateTx mocks base method
func (m *MockUseCases) SendGoQuorumPrivateTx() usecases.SendGoQuorumPrivateTxUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendGoQuorumPrivateTx")
	ret0, _ := ret[0].(usecases.SendGoQuorumPrivateTxUseCase)
	return ret0
}

// SendGoQuorumPrivateTx indicates an expected call of SendGoQuorumPrivateTx
func (mr *MockUseCasesMockRecorder) SendGoQuorumPrivateTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGoQuorumPrivateTx", reflect.TypeOf((*MockUseCases)(nil).SendGoQuorumPrivateTx))
}

// SendGoQuorumMarkingTx mocks base method
func (m *MockUseCases) SendGoQuorumMarkingTx() usecases.SendGoQuorumMarkingTxUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendGoQuorumMarkingTx")
	ret0, _ := ret[0].(usecases.SendGoQuorumMarkingTxUseCase)
	return ret0
}

// SendGoQuorumMarkingTx indicates an expected call of SendGoQuorumMarkingTx
func (mr *MockUseCasesMockRecorder) SendGoQuorumMarkingTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGoQuorumMarkingTx", reflect.TypeOf((*MockUseCases)(nil).SendGoQuorumMarkingTx))
}