// Code generated by MockGen. DO NOT EDIT.
// Source: jobs.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	entities "github.com/consensys/orchestrate/src/entities"
	usecases "github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

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

// Execute mocks base method
func (m *MockPendingJob) Execute(ctx context.Context, job *entities.Job, msg *entities.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job, msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockPendingJobMockRecorder) Execute(ctx, job, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockPendingJob)(nil).Execute), ctx, job, msg)
}

// MockFailedJob is a mock of FailedJob interface
type MockFailedJob struct {
	ctrl     *gomock.Controller
	recorder *MockFailedJobMockRecorder
}

// MockFailedJobMockRecorder is the mock recorder for MockFailedJob
type MockFailedJobMockRecorder struct {
	mock *MockFailedJob
}

// NewMockFailedJob creates a new mock instance
func NewMockFailedJob(ctrl *gomock.Controller) *MockFailedJob {
	mock := &MockFailedJob{ctrl: ctrl}
	mock.recorder = &MockFailedJobMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFailedJob) EXPECT() *MockFailedJobMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockFailedJob) Execute(ctx context.Context, job *entities.Job, errMsg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job, errMsg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockFailedJobMockRecorder) Execute(ctx, job, errMsg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockFailedJob)(nil).Execute), ctx, job, errMsg)
}

// MockMinedJob is a mock of MinedJob interface
type MockMinedJob struct {
	ctrl     *gomock.Controller
	recorder *MockMinedJobMockRecorder
}

// MockMinedJobMockRecorder is the mock recorder for MockMinedJob
type MockMinedJobMockRecorder struct {
	mock *MockMinedJob
}

// NewMockMinedJob creates a new mock instance
func NewMockMinedJob(ctrl *gomock.Controller) *MockMinedJob {
	mock := &MockMinedJob{ctrl: ctrl}
	mock.recorder = &MockMinedJobMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMinedJob) EXPECT() *MockMinedJobMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockMinedJob) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockMinedJobMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockMinedJob)(nil).Execute), ctx, job)
}

// MockCompletedJob is a mock of CompletedJob interface
type MockCompletedJob struct {
	ctrl     *gomock.Controller
	recorder *MockCompletedJobMockRecorder
}

// MockCompletedJobMockRecorder is the mock recorder for MockCompletedJob
type MockCompletedJobMockRecorder struct {
	mock *MockCompletedJob
}

// NewMockCompletedJob creates a new mock instance
func NewMockCompletedJob(ctrl *gomock.Controller) *MockCompletedJob {
	mock := &MockCompletedJob{ctrl: ctrl}
	mock.recorder = &MockCompletedJobMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCompletedJob) EXPECT() *MockCompletedJobMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockCompletedJob) Execute(ctx context.Context, job *entities.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockCompletedJobMockRecorder) Execute(ctx, job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockCompletedJob)(nil).Execute), ctx, job)
}

// MockRetryJob is a mock of RetryJob interface
type MockRetryJob struct {
	ctrl     *gomock.Controller
	recorder *MockRetryJobMockRecorder
}

// MockRetryJobMockRecorder is the mock recorder for MockRetryJob
type MockRetryJobMockRecorder struct {
	mock *MockRetryJob
}

// NewMockRetryJob creates a new mock instance
func NewMockRetryJob(ctrl *gomock.Controller) *MockRetryJob {
	mock := &MockRetryJob{ctrl: ctrl}
	mock.recorder = &MockRetryJobMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRetryJob) EXPECT() *MockRetryJobMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockRetryJob) Execute(ctx context.Context, job *entities.Job, lastChildUUID string, nChildren int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job, lastChildUUID, nChildren)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockRetryJobMockRecorder) Execute(ctx, job, lastChildUUID, nChildren interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockRetryJob)(nil).Execute), ctx, job, lastChildUUID, nChildren)
}

// MockJobUseCases is a mock of JobUseCases interface
type MockJobUseCases struct {
	ctrl     *gomock.Controller
	recorder *MockJobUseCasesMockRecorder
}

// MockJobUseCasesMockRecorder is the mock recorder for MockJobUseCases
type MockJobUseCasesMockRecorder struct {
	mock *MockJobUseCases
}

// NewMockJobUseCases creates a new mock instance
func NewMockJobUseCases(ctrl *gomock.Controller) *MockJobUseCases {
	mock := &MockJobUseCases{ctrl: ctrl}
	mock.recorder = &MockJobUseCasesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJobUseCases) EXPECT() *MockJobUseCasesMockRecorder {
	return m.recorder
}

// PendingJobUseCase mocks base method
func (m *MockJobUseCases) PendingJobUseCase() usecases.PendingJob {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PendingJobUseCase")
	ret0, _ := ret[0].(usecases.PendingJob)
	return ret0
}

// PendingJobUseCase indicates an expected call of PendingJobUseCase
func (mr *MockJobUseCasesMockRecorder) PendingJobUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PendingJobUseCase", reflect.TypeOf((*MockJobUseCases)(nil).PendingJobUseCase))
}

// FailedJobUseCase mocks base method
func (m *MockJobUseCases) FailedJobUseCase() usecases.FailedJob {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FailedJobUseCase")
	ret0, _ := ret[0].(usecases.FailedJob)
	return ret0
}

// FailedJobUseCase indicates an expected call of FailedJobUseCase
func (mr *MockJobUseCasesMockRecorder) FailedJobUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FailedJobUseCase", reflect.TypeOf((*MockJobUseCases)(nil).FailedJobUseCase))
}

// MinedJobUseCase mocks base method
func (m *MockJobUseCases) MinedJobUseCase() usecases.MinedJob {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MinedJobUseCase")
	ret0, _ := ret[0].(usecases.MinedJob)
	return ret0
}

// MinedJobUseCase indicates an expected call of MinedJobUseCase
func (mr *MockJobUseCasesMockRecorder) MinedJobUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MinedJobUseCase", reflect.TypeOf((*MockJobUseCases)(nil).MinedJobUseCase))
}

// RetryJobUseCase mocks base method
func (m *MockJobUseCases) RetryJobUseCase() usecases.RetryJob {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetryJobUseCase")
	ret0, _ := ret[0].(usecases.RetryJob)
	return ret0
}

// RetryJobUseCase indicates an expected call of RetryJobUseCase
func (mr *MockJobUseCasesMockRecorder) RetryJobUseCase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetryJobUseCase", reflect.TypeOf((*MockJobUseCases)(nil).RetryJobUseCase))
}
