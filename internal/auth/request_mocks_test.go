// Code generated by MockGen. DO NOT EDIT.
// Source: ./request.go

// Package identity is a generated GoMock package.
package auth

import (
	context "context"
	reflect "reflect"

	common "github.com/ethereum/go-ethereum/common"
	gomock "go.uber.org/mock/gomock"
)

// MockIdentityService is a mock of IdentityService interface.
type MockIdentityService struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityServiceMockRecorder
}

// MockIdentityServiceMockRecorder is the mock recorder for MockIdentityService.
type MockIdentityServiceMockRecorder struct {
	mock *MockIdentityService
}

// NewMockIdentityService creates a new mock instance.
func NewMockIdentityService(ctrl *gomock.Controller) *MockIdentityService {
	mock := &MockIdentityService{ctrl: ctrl}
	mock.recorder = &MockIdentityServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIdentityService) EXPECT() *MockIdentityServiceMockRecorder {
	return m.recorder
}

// AftermarketDevice mocks base method.
func (m *MockIdentityService) AftermarketDevice(ctx context.Context, address *common.Address, tokenID *int, serial *string) (*ManufacturerTokenID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AftermarketDevice", ctx, address, tokenID, serial)
	ret0, _ := ret[0].(*ManufacturerTokenID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AftermarketDevice indicates an expected call of AftermarketDevice.
func (mr *MockIdentityServiceMockRecorder) AftermarketDevice(ctx, address, tokenID, serial interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AftermarketDevice", reflect.TypeOf((*MockIdentityService)(nil).AftermarketDevice), ctx, address, tokenID, serial)
}
