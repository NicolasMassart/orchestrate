// Code generated by MockGen. DO NOT EDIT.
// Source: subscriptions.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	multitenancy "github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	entities "github.com/consensys/orchestrate/src/entities"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

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

// Get mocks base method
func (m *MockSubscriptionUseCases) Get() usecases.GetSubscriptionUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].(usecases.GetSubscriptionUseCase)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockSubscriptionUseCasesMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSubscriptionUseCases)(nil).Get))
}

// Create mocks base method
func (m *MockSubscriptionUseCases) Create() usecases.CreateSubscriptionUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create")
	ret0, _ := ret[0].(usecases.CreateSubscriptionUseCase)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockSubscriptionUseCasesMockRecorder) Create() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSubscriptionUseCases)(nil).Create))
}

// Update mocks base method
func (m *MockSubscriptionUseCases) Update() usecases.UpdateSubscriptionUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update")
	ret0, _ := ret[0].(usecases.UpdateSubscriptionUseCase)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockSubscriptionUseCasesMockRecorder) Update() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSubscriptionUseCases)(nil).Update))
}

// Search mocks base method
func (m *MockSubscriptionUseCases) Search() usecases.SearchSubscriptionUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search")
	ret0, _ := ret[0].(usecases.SearchSubscriptionUseCase)
	return ret0
}

// Search indicates an expected call of Search
func (mr *MockSubscriptionUseCasesMockRecorder) Search() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockSubscriptionUseCases)(nil).Search))
}

// Delete mocks base method
func (m *MockSubscriptionUseCases) Delete() usecases.DeleteSubscriptionUseCase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete")
	ret0, _ := ret[0].(usecases.DeleteSubscriptionUseCase)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockSubscriptionUseCasesMockRecorder) Delete() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSubscriptionUseCases)(nil).Delete))
}

// MockGetSubscriptionUseCase is a mock of GetSubscriptionUseCase interface
type MockGetSubscriptionUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockGetSubscriptionUseCaseMockRecorder
}

// MockGetSubscriptionUseCaseMockRecorder is the mock recorder for MockGetSubscriptionUseCase
type MockGetSubscriptionUseCaseMockRecorder struct {
	mock *MockGetSubscriptionUseCase
}

// NewMockGetSubscriptionUseCase creates a new mock instance
func NewMockGetSubscriptionUseCase(ctrl *gomock.Controller) *MockGetSubscriptionUseCase {
	mock := &MockGetSubscriptionUseCase{ctrl: ctrl}
	mock.recorder = &MockGetSubscriptionUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGetSubscriptionUseCase) EXPECT() *MockGetSubscriptionUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockGetSubscriptionUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, uuid, userInfo)
	ret0, _ := ret[0].(*entities.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockGetSubscriptionUseCaseMockRecorder) Execute(ctx, uuid, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockGetSubscriptionUseCase)(nil).Execute), ctx, uuid, userInfo)
}

// MockCreateSubscriptionUseCase is a mock of CreateSubscriptionUseCase interface
type MockCreateSubscriptionUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockCreateSubscriptionUseCaseMockRecorder
}

// MockCreateSubscriptionUseCaseMockRecorder is the mock recorder for MockCreateSubscriptionUseCase
type MockCreateSubscriptionUseCaseMockRecorder struct {
	mock *MockCreateSubscriptionUseCase
}

// NewMockCreateSubscriptionUseCase creates a new mock instance
func NewMockCreateSubscriptionUseCase(ctrl *gomock.Controller) *MockCreateSubscriptionUseCase {
	mock := &MockCreateSubscriptionUseCase{ctrl: ctrl}
	mock.recorder = &MockCreateSubscriptionUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCreateSubscriptionUseCase) EXPECT() *MockCreateSubscriptionUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockCreateSubscriptionUseCase) Execute(ctx context.Context, Subscription *entities.Subscription, chainName, eventStreamName string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, Subscription, chainName, eventStreamName, userInfo)
	ret0, _ := ret[0].(*entities.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockCreateSubscriptionUseCaseMockRecorder) Execute(ctx, Subscription, chainName, eventStreamName, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockCreateSubscriptionUseCase)(nil).Execute), ctx, Subscription, chainName, eventStreamName, userInfo)
}

// MockUpdateSubscriptionUseCase is a mock of UpdateSubscriptionUseCase interface
type MockUpdateSubscriptionUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockUpdateSubscriptionUseCaseMockRecorder
}

// MockUpdateSubscriptionUseCaseMockRecorder is the mock recorder for MockUpdateSubscriptionUseCase
type MockUpdateSubscriptionUseCaseMockRecorder struct {
	mock *MockUpdateSubscriptionUseCase
}

// NewMockUpdateSubscriptionUseCase creates a new mock instance
func NewMockUpdateSubscriptionUseCase(ctrl *gomock.Controller) *MockUpdateSubscriptionUseCase {
	mock := &MockUpdateSubscriptionUseCase{ctrl: ctrl}
	mock.recorder = &MockUpdateSubscriptionUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUpdateSubscriptionUseCase) EXPECT() *MockUpdateSubscriptionUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockUpdateSubscriptionUseCase) Execute(ctx context.Context, Subscription *entities.Subscription, eventStreamName string, userInfo *multitenancy.UserInfo) (*entities.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, Subscription, eventStreamName, userInfo)
	ret0, _ := ret[0].(*entities.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockUpdateSubscriptionUseCaseMockRecorder) Execute(ctx, Subscription, eventStreamName, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockUpdateSubscriptionUseCase)(nil).Execute), ctx, Subscription, eventStreamName, userInfo)
}

// MockSearchSubscriptionUseCase is a mock of SearchSubscriptionUseCase interface
type MockSearchSubscriptionUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockSearchSubscriptionUseCaseMockRecorder
}

// MockSearchSubscriptionUseCaseMockRecorder is the mock recorder for MockSearchSubscriptionUseCase
type MockSearchSubscriptionUseCaseMockRecorder struct {
	mock *MockSearchSubscriptionUseCase
}

// NewMockSearchSubscriptionUseCase creates a new mock instance
func NewMockSearchSubscriptionUseCase(ctrl *gomock.Controller) *MockSearchSubscriptionUseCase {
	mock := &MockSearchSubscriptionUseCase{ctrl: ctrl}
	mock.recorder = &MockSearchSubscriptionUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSearchSubscriptionUseCase) EXPECT() *MockSearchSubscriptionUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockSearchSubscriptionUseCase) Execute(ctx context.Context, filters *entities.SubscriptionFilters, userInfo *multitenancy.UserInfo) ([]*entities.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, filters, userInfo)
	ret0, _ := ret[0].([]*entities.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute
func (mr *MockSearchSubscriptionUseCaseMockRecorder) Execute(ctx, filters, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockSearchSubscriptionUseCase)(nil).Execute), ctx, filters, userInfo)
}

// MockDeleteSubscriptionUseCase is a mock of DeleteSubscriptionUseCase interface
type MockDeleteSubscriptionUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockDeleteSubscriptionUseCaseMockRecorder
}

// MockDeleteSubscriptionUseCaseMockRecorder is the mock recorder for MockDeleteSubscriptionUseCase
type MockDeleteSubscriptionUseCaseMockRecorder struct {
	mock *MockDeleteSubscriptionUseCase
}

// NewMockDeleteSubscriptionUseCase creates a new mock instance
func NewMockDeleteSubscriptionUseCase(ctrl *gomock.Controller) *MockDeleteSubscriptionUseCase {
	mock := &MockDeleteSubscriptionUseCase{ctrl: ctrl}
	mock.recorder = &MockDeleteSubscriptionUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeleteSubscriptionUseCase) EXPECT() *MockDeleteSubscriptionUseCaseMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockDeleteSubscriptionUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", ctx, uuid, userInfo)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockDeleteSubscriptionUseCaseMockRecorder) Execute(ctx, uuid, userInfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockDeleteSubscriptionUseCase)(nil).Execute), ctx, uuid, userInfo)
}