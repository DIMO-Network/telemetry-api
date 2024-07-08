// Code generated by MockGen. DO NOT EDIT.
// Source: ./repositories.go
//
// Generated by this command:
//
//	mockgen -source=./repositories.go -destination=repositories_mocks_test.go -package=repositories_test
//

// Package repositories_test is a generated GoMock package.
package repositories_test

import (
	context "context"
	reflect "reflect"

	vss "github.com/DIMO-Network/model-garage/pkg/vss"
	model "github.com/DIMO-Network/telemetry-api/internal/graph/model"
	gomock "go.uber.org/mock/gomock"
)

// MockCHService is a mock of CHService interface.
type MockCHService struct {
	ctrl     *gomock.Controller
	recorder *MockCHServiceMockRecorder
}

// MockCHServiceMockRecorder is the mock recorder for MockCHService.
type MockCHServiceMockRecorder struct {
	mock *MockCHService
}

// NewMockCHService creates a new mock instance.
func NewMockCHService(ctrl *gomock.Controller) *MockCHService {
	mock := &MockCHService{ctrl: ctrl}
	mock.recorder = &MockCHServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCHService) EXPECT() *MockCHServiceMockRecorder {
	return m.recorder
}

// GetAggregatedSignals mocks base method.
func (m *MockCHService) GetAggregatedSignals(ctx context.Context, aggArgs *model.AggregatedSignalArgs) ([]*model.AggSignal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAggregatedSignals", ctx, aggArgs)
	ret0, _ := ret[0].([]*model.AggSignal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAggregatedSignals indicates an expected call of GetAggregatedSignals.
func (mr *MockCHServiceMockRecorder) GetAggregatedSignals(ctx, aggArgs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAggregatedSignals", reflect.TypeOf((*MockCHService)(nil).GetAggregatedSignals), ctx, aggArgs)
}

// GetLatestSignals mocks base method.
func (m *MockCHService) GetLatestSignals(ctx context.Context, latestArgs *model.LatestSignalsArgs) ([]*vss.Signal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLatestSignals", ctx, latestArgs)
	ret0, _ := ret[0].([]*vss.Signal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLatestSignals indicates an expected call of GetLatestSignals.
func (mr *MockCHServiceMockRecorder) GetLatestSignals(ctx, latestArgs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestSignals", reflect.TypeOf((*MockCHService)(nil).GetLatestSignals), ctx, latestArgs)
}
