// Code generated by MockGen. DO NOT EDIT.
// Source: contracts.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	entities "github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockRegisterDeployedContract is a mock of RegisterDeployedContract interface
type MockRegisterDeployedContract struct {
	ctrl     *gomock.Controller
	recorder *MockRegisterDeployedContractMockRecorder
}

// MockRegisterDeployedContractMockRecorder is the mock recorder for MockRegisterDeployedContract
type MockRegisterDeployedContractMockRecorder struct {
	mock *MockRegisterDeployedContract
}

// NewMockRegisterDeployedContract creates a new mock instance
func NewMockRegisterDeployedContract(ctrl *gomock.Controller) *MockRegisterDeployedContract {
	mock := &MockRegisterDeployedContract{ctrl: ctrl}
	mock.recorder = &MockRegisterDeployedContractMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRegisterDeployedContract) EXPECT() *MockRegisterDeployedContractMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockRegisterDeployedContract) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockRegisterDeployedContractMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockRegisterDeployedContract)(nil).Execute), ctx, job)
}

// MockContractsUseCases is a mock of ContractsUseCases interface
type MockContractsUseCases struct {
	ctrl     *gomock.Controller
	recorder *MockContractsUseCasesMockRecorder
}

// MockContractsUseCasesMockRecorder is the mock recorder for MockContractsUseCases
type MockContractsUseCasesMockRecorder struct {
	mock *MockContractsUseCases
}

// NewMockContractsUseCases creates a new mock instance
func NewMockContractsUseCases(ctrl *gomock.Controller) *MockContractsUseCases {
	mock := &MockContractsUseCases{ctrl: ctrl}
	mock.recorder = &MockContractsUseCasesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockContractsUseCases) EXPECT() *MockContractsUseCasesMockRecorder {
	return m.recorder
}

// RegisterDeployedContractUseCase mocks base method
func (m *MockContractsUseCases) RegisterDeployedContractUseCase() usecases.RegisterDeployedContract {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterDeployedContractUseCase")
	ret0, _ := ret[0].(usecases.RegisterDeployedContract)
	return ret0
}

// RegisterDeployedContractUseCase indicates an expected call of RegisterDeployedContractUseCase
func (mr *MockContractsUseCasesMockRecorder) RegisterDeployedContractUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterDeployedContractUseCase", reflect.TypeOf((*MockContractsUseCases)(nil).RegisterDeployedContractUseCase))
}