// Code generated by MockGen. DO NOT EDIT.
// Source: server.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	types "github.com/Edbeer/paymentapi/types"
	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// CreateAccount mocks base method.
func (m *MockStorage) CreateAccount(ctx context.Context, reqAcc *types.RequestCreate) (*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccount", ctx, reqAcc)
	ret0, _ := ret[0].(*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAccount indicates an expected call of CreateAccount.
func (mr *MockStorageMockRecorder) CreateAccount(ctx, reqAcc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccount", reflect.TypeOf((*MockStorage)(nil).CreateAccount), ctx, reqAcc)
}

// DeleteAccount mocks base method.
func (m *MockStorage) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAccount", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAccount indicates an expected call of DeleteAccount.
func (mr *MockStorageMockRecorder) DeleteAccount(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAccount", reflect.TypeOf((*MockStorage)(nil).DeleteAccount), ctx, id)
}

// DepositAccount mocks base method.
func (m *MockStorage) DepositAccount(ctx context.Context, reqDep *types.RequestDeposit) (*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DepositAccount", ctx, reqDep)
	ret0, _ := ret[0].(*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DepositAccount indicates an expected call of DepositAccount.
func (mr *MockStorageMockRecorder) DepositAccount(ctx, reqDep interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DepositAccount", reflect.TypeOf((*MockStorage)(nil).DepositAccount), ctx, reqDep)
}

// GetAccount mocks base method.
func (m *MockStorage) GetAccount(ctx context.Context) ([]*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccount", ctx)
	ret0, _ := ret[0].([]*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccount indicates an expected call of GetAccount.
func (mr *MockStorageMockRecorder) GetAccount(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockStorage)(nil).GetAccount), ctx)
}

// GetAccountByCard mocks base method.
func (m *MockStorage) GetAccountByCard(ctx context.Context, card string) (*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountByCard", ctx, card)
	ret0, _ := ret[0].(*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountByCard indicates an expected call of GetAccountByCard.
func (mr *MockStorageMockRecorder) GetAccountByCard(ctx, card interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountByCard", reflect.TypeOf((*MockStorage)(nil).GetAccountByCard), ctx, card)
}

// GetAccountByID mocks base method.
func (m *MockStorage) GetAccountByID(ctx context.Context, id uuid.UUID) (*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountByID", ctx, id)
	ret0, _ := ret[0].(*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountByID indicates an expected call of GetAccountByID.
func (mr *MockStorageMockRecorder) GetAccountByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountByID", reflect.TypeOf((*MockStorage)(nil).GetAccountByID), ctx, id)
}

// GetAccountStatement mocks base method.
func (m *MockStorage) GetAccountStatement(ctx context.Context, id uuid.UUID) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountStatement", ctx, id)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountStatement indicates an expected call of GetAccountStatement.
func (mr *MockStorageMockRecorder) GetAccountStatement(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountStatement", reflect.TypeOf((*MockStorage)(nil).GetAccountStatement), ctx, id)
}

// GetPaymentByID mocks base method.
func (m *MockStorage) GetPaymentByID(ctx context.Context, id uuid.UUID) (*types.Payment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaymentByID", ctx, id)
	ret0, _ := ret[0].(*types.Payment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPaymentByID indicates an expected call of GetPaymentByID.
func (mr *MockStorageMockRecorder) GetPaymentByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaymentByID", reflect.TypeOf((*MockStorage)(nil).GetPaymentByID), ctx, id)
}

// SaveBalance mocks base method.
func (m *MockStorage) SaveBalance(ctx context.Context, tx *sql.Tx, account *types.Account, balance, bmoney uint64) (*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveBalance", ctx, tx, account, balance, bmoney)
	ret0, _ := ret[0].(*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveBalance indicates an expected call of SaveBalance.
func (mr *MockStorageMockRecorder) SaveBalance(ctx, tx, account, balance, bmoney interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveBalance", reflect.TypeOf((*MockStorage)(nil).SaveBalance), ctx, tx, account, balance, bmoney)
}

// SavePayment mocks base method.
func (m *MockStorage) SavePayment(ctx context.Context, tx *sql.Tx, payment *types.Payment) (*types.Payment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SavePayment", ctx, tx, payment)
	ret0, _ := ret[0].(*types.Payment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SavePayment indicates an expected call of SavePayment.
func (mr *MockStorageMockRecorder) SavePayment(ctx, tx, payment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SavePayment", reflect.TypeOf((*MockStorage)(nil).SavePayment), ctx, tx, payment)
}

// UpdateAccount mocks base method.
func (m *MockStorage) UpdateAccount(ctx context.Context, reqUp *types.RequestUpdate, id uuid.UUID) (*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccount", ctx, reqUp, id)
	ret0, _ := ret[0].(*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateAccount indicates an expected call of UpdateAccount.
func (mr *MockStorageMockRecorder) UpdateAccount(ctx, reqUp, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccount", reflect.TypeOf((*MockStorage)(nil).UpdateAccount), ctx, reqUp, id)
}

// UpdateStatement mocks base method.
func (m *MockStorage) UpdateStatement(ctx context.Context, tx *sql.Tx, id, paymentId uuid.UUID) (*types.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatement", ctx, tx, id, paymentId)
	ret0, _ := ret[0].(*types.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateStatement indicates an expected call of UpdateStatement.
func (mr *MockStorageMockRecorder) UpdateStatement(ctx, tx, id, paymentId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatement", reflect.TypeOf((*MockStorage)(nil).UpdateStatement), ctx, tx, id, paymentId)
}

// MockRedisStorage is a mock of RedisStorage interface.
type MockRedisStorage struct {
	ctrl     *gomock.Controller
	recorder *MockRedisStorageMockRecorder
}

// MockRedisStorageMockRecorder is the mock recorder for MockRedisStorage.
type MockRedisStorageMockRecorder struct {
	mock *MockRedisStorage
}

// NewMockRedisStorage creates a new mock instance.
func NewMockRedisStorage(ctrl *gomock.Controller) *MockRedisStorage {
	mock := &MockRedisStorage{ctrl: ctrl}
	mock.recorder = &MockRedisStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRedisStorage) EXPECT() *MockRedisStorageMockRecorder {
	return m.recorder
}

// CreateSession mocks base method.
func (m *MockRedisStorage) CreateSession(ctx context.Context, session *types.Session, expire int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSession", ctx, session, expire)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSession indicates an expected call of CreateSession.
func (mr *MockRedisStorageMockRecorder) CreateSession(ctx, session, expire interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSession", reflect.TypeOf((*MockRedisStorage)(nil).CreateSession), ctx, session, expire)
}

// DeleteSession mocks base method.
func (m *MockRedisStorage) DeleteSession(ctx context.Context, refreshToken string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSession", ctx, refreshToken)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSession indicates an expected call of DeleteSession.
func (mr *MockRedisStorageMockRecorder) DeleteSession(ctx, refreshToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSession", reflect.TypeOf((*MockRedisStorage)(nil).DeleteSession), ctx, refreshToken)
}

// GetUserID mocks base method.
func (m *MockRedisStorage) GetUserID(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserID", ctx, refreshToken)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserID indicates an expected call of GetUserID.
func (mr *MockRedisStorageMockRecorder) GetUserID(ctx, refreshToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserID", reflect.TypeOf((*MockRedisStorage)(nil).GetUserID), ctx, refreshToken)
}
