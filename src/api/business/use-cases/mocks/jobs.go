// Code generated by MockGen. DO NOT EDIT.
// Source: jobs.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	multitenancy "github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	entities "github.com/consensys/orchestrate/src/entities"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

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

// Create mocks base method
func (m *MockJobUseCases) Create() usecases.CreateJobUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create")
	ret0, _ := ret[0].(usecases.CreateJobUseCase)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockJobUseCasesMockRecorder) Create() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockJobUseCases)(nil).Create))
}

// Get mocks base method
func (m *MockJobUseCases) Get() usecases.GetJobUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].(usecases.GetJobUseCase)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockJobUseCasesMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockJobUseCases)(nil).Get))
}

// Start mocks base method
func (m *MockJobUseCases) Start() usecases.StartJobUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(usecases.StartJobUseCase)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockJobUseCasesMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockJobUseCases)(nil).Start))
}

// ResendTx mocks base method
func (m *MockJobUseCases) ResendTx() usecases.ResendJobTxUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResendTx")
	ret0, _ := ret[0].(usecases.ResendJobTxUseCase)
	return ret0
}

// ResendTx indicates an expected call of ResendTx
func (mr *MockJobUseCasesMockRecorder) ResendTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResendTx", reflect.TypeOf((*MockJobUseCases)(nil).ResendTx))
}

// RetryTx mocks base method
func (m *MockJobUseCases) RetryTx() usecases.RetryJobTxUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetryTx")
	ret0, _ := ret[0].(usecases.RetryJobTxUseCase)
	return ret0
}

// RetryTx indicates an expected call of RetryTx
func (mr *MockJobUseCasesMockRecorder) RetryTx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetryTx", reflect.TypeOf((*MockJobUseCases)(nil).RetryTx))
}

// Update mocks base method
func (m *MockJobUseCases) Update() usecases.UpdateJobUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update")
	ret0, _ := ret[0].(usecases.UpdateJobUseCase)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockJobUseCasesMockRecorder) Update() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockJobUseCases)(nil).Update))
}

// Search mocks base method
func (m *MockJobUseCases) Search() usecases.SearchJobsUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search")
	ret0, _ := ret[0].(usecases.SearchJobsUseCase)
	return ret0
}

// Search indicates an expected call of Search
func (mr *MockJobUseCasesMockRecorder) Search() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockJobUseCases)(nil).Search))
}

// MockCreateJobUseCase is a mock of CreateJobUseCase interface
type MockCreateJobUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockCreateJobUseCaseMockRecorder
}

// MockCreateJobUseCaseMockRecorder is the mock recorder for MockCreateJobUseCase
type MockCreateJobUseCaseMockRecorder struct {
	mock *MockCreateJobUseCase
}

// NewMockCreateJobUseCase creates a new mock instance
func NewMockCreateJobUseCase(ctrl *gomock.Controller) *MockCreateJobUseCase {
	mock := &MockCreateJobUseCase{ctrl: ctrl}
	mock.recorder = &MockCreateJobUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCreateJobUseCase) EXPECT() *MockCreateJobUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockCreateJobUseCase) Execute(ctx context.Context, job *entities.Job, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, job, userInfo)
	ret0, _ := ret[0].(*entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockCreateJobUseCaseMockRecorder) Execute(ctx, job, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockCreateJobUseCase)(nil).Execute), ctx, job, userInfo)
}

// MockGetJobUseCase is a mock of GetJobUseCase interface
type MockGetJobUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockGetJobUseCaseMockRecorder
}

// MockGetJobUseCaseMockRecorder is the mock recorder for MockGetJobUseCase
type MockGetJobUseCaseMockRecorder struct {
	mock *MockGetJobUseCase
}

// NewMockGetJobUseCase creates a new mock instance
func NewMockGetJobUseCase(ctrl *gomock.Controller) *MockGetJobUseCase {
	mock := &MockGetJobUseCase{ctrl: ctrl}
	mock.recorder = &MockGetJobUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGetJobUseCase) EXPECT() *MockGetJobUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockGetJobUseCase) Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, jobUUID, userInfo)
	ret0, _ := ret[0].(*entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockGetJobUseCaseMockRecorder) Execute(ctx, jobUUID, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockGetJobUseCase)(nil).Execute), ctx, jobUUID, userInfo)
}

// MockSearchJobsUseCase is a mock of SearchJobsUseCase interface
type MockSearchJobsUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockSearchJobsUseCaseMockRecorder
}

// MockSearchJobsUseCaseMockRecorder is the mock recorder for MockSearchJobsUseCase
type MockSearchJobsUseCaseMockRecorder struct {
	mock *MockSearchJobsUseCase
}

// NewMockSearchJobsUseCase creates a new mock instance
func NewMockSearchJobsUseCase(ctrl *gomock.Controller) *MockSearchJobsUseCase {
	mock := &MockSearchJobsUseCase{ctrl: ctrl}
	mock.recorder = &MockSearchJobsUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSearchJobsUseCase) EXPECT() *MockSearchJobsUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockSearchJobsUseCase) Execute(ctx context.Context, filters *entities.JobFilters, userInfo *multitenancy.UserInfo) ([]*entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, filters, userInfo)
	ret0, _ := ret[0].([]*entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockSearchJobsUseCaseMockRecorder) Execute(ctx, filters, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockSearchJobsUseCase)(nil).Execute), ctx, filters, userInfo)
}

// MockStartJobUseCase is a mock of StartJobUseCase interface
type MockStartJobUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockStartJobUseCaseMockRecorder
}

// MockStartJobUseCaseMockRecorder is the mock recorder for MockStartJobUseCase
type MockStartJobUseCaseMockRecorder struct {
	mock *MockStartJobUseCase
}

// NewMockStartJobUseCase creates a new mock instance
func NewMockStartJobUseCase(ctrl *gomock.Controller) *MockStartJobUseCase {
	mock := &MockStartJobUseCase{ctrl: ctrl}
	mock.recorder = &MockStartJobUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStartJobUseCase) EXPECT() *MockStartJobUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockStartJobUseCase) Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, jobUUID, userInfo)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockStartJobUseCaseMockRecorder) Execute(ctx, jobUUID, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockStartJobUseCase)(nil).Execute), ctx, jobUUID, userInfo)
}

// MockStartNextJobUseCase is a mock of StartNextJobUseCase interface
type MockStartNextJobUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockStartNextJobUseCaseMockRecorder
}

// MockStartNextJobUseCaseMockRecorder is the mock recorder for MockStartNextJobUseCase
type MockStartNextJobUseCaseMockRecorder struct {
	mock *MockStartNextJobUseCase
}

// NewMockStartNextJobUseCase creates a new mock instance
func NewMockStartNextJobUseCase(ctrl *gomock.Controller) *MockStartNextJobUseCase {
	mock := &MockStartNextJobUseCase{ctrl: ctrl}
	mock.recorder = &MockStartNextJobUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStartNextJobUseCase) EXPECT() *MockStartNextJobUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockStartNextJobUseCase) Execute(ctx context.Context, prevJobUUID string, userInfo *multitenancy.UserInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, prevJobUUID, userInfo)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockStartNextJobUseCaseMockRecorder) Execute(ctx, prevJobUUID, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockStartNextJobUseCase)(nil).Execute), ctx, prevJobUUID, userInfo)
}

// MockUpdateJobUseCase is a mock of UpdateJobUseCase interface
type MockUpdateJobUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockUpdateJobUseCaseMockRecorder
}

// MockUpdateJobUseCaseMockRecorder is the mock recorder for MockUpdateJobUseCase
type MockUpdateJobUseCaseMockRecorder struct {
	mock *MockUpdateJobUseCase
}

// NewMockUpdateJobUseCase creates a new mock instance
func NewMockUpdateJobUseCase(ctrl *gomock.Controller) *MockUpdateJobUseCase {
	mock := &MockUpdateJobUseCase{ctrl: ctrl}
	mock.recorder = &MockUpdateJobUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUpdateJobUseCase) EXPECT() *MockUpdateJobUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockUpdateJobUseCase) Execute(ctx context.Context, jobEntity *entities.Job, nextStatus entities.JobStatus, logMessage string, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, jobEntity, nextStatus, logMessage, userInfo)
	ret0, _ := ret[0].(*entities.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockUpdateJobUseCaseMockRecorder) Execute(ctx, jobEntity, nextStatus, logMessage, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockUpdateJobUseCase)(nil).Execute), ctx, jobEntity, nextStatus, logMessage, userInfo)
}

// MockResendJobTxUseCase is a mock of ResendJobTxUseCase interface
type MockResendJobTxUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockResendJobTxUseCaseMockRecorder
}

// MockResendJobTxUseCaseMockRecorder is the mock recorder for MockResendJobTxUseCase
type MockResendJobTxUseCaseMockRecorder struct {
	mock *MockResendJobTxUseCase
}

// NewMockResendJobTxUseCase creates a new mock instance
func NewMockResendJobTxUseCase(ctrl *gomock.Controller) *MockResendJobTxUseCase {
	mock := &MockResendJobTxUseCase{ctrl: ctrl}
	mock.recorder = &MockResendJobTxUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockResendJobTxUseCase) EXPECT() *MockResendJobTxUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockResendJobTxUseCase) Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, jobUUID, userInfo)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockResendJobTxUseCaseMockRecorder) Execute(ctx, jobUUID, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockResendJobTxUseCase)(nil).Execute), ctx, jobUUID, userInfo)
}

// MockRetryJobTxUseCase is a mock of RetryJobTxUseCase interface
type MockRetryJobTxUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockRetryJobTxUseCaseMockRecorder
}

// MockRetryJobTxUseCaseMockRecorder is the mock recorder for MockRetryJobTxUseCase
type MockRetryJobTxUseCaseMockRecorder struct {
	mock *MockRetryJobTxUseCase
}

// NewMockRetryJobTxUseCase creates a new mock instance
func NewMockRetryJobTxUseCase(ctrl *gomock.Controller) *MockRetryJobTxUseCase {
	mock := &MockRetryJobTxUseCase{ctrl: ctrl}
	mock.recorder = &MockRetryJobTxUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRetryJobTxUseCase) EXPECT() *MockRetryJobTxUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockRetryJobTxUseCase) Execute(ctx context.Context, jobUUID string, gasIncrement float64, data hexutil.Bytes, userInfo *multitenancy.UserInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, jobUUID, gasIncrement, data, userInfo)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockRetryJobTxUseCaseMockRecorder) Execute(ctx, jobUUID, gasIncrement, data, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockRetryJobTxUseCase)(nil).Execute), ctx, jobUUID, gasIncrement, data, userInfo)
}
