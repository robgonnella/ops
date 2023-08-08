// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/robgonnella/ops/internal/discovery (interfaces: DetailScanner,Scanner)

// Package mock_discovery is a generated GoMock package.
package mock_discovery

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	discovery "github.com/robgonnella/ops/internal/discovery"
)

// MockDetailScanner is a mock of DetailScanner interface.
type MockDetailScanner struct {
	ctrl     *gomock.Controller
	recorder *MockDetailScannerMockRecorder
}

// MockDetailScannerMockRecorder is the mock recorder for MockDetailScanner.
type MockDetailScannerMockRecorder struct {
	mock *MockDetailScanner
}

// NewMockDetailScanner creates a new mock instance.
func NewMockDetailScanner(ctrl *gomock.Controller) *MockDetailScanner {
	mock := &MockDetailScanner{ctrl: ctrl}
	mock.recorder = &MockDetailScannerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDetailScanner) EXPECT() *MockDetailScannerMockRecorder {
	return m.recorder
}

// GetServerDetails mocks base method.
func (m *MockDetailScanner) GetServerDetails(arg0 context.Context, arg1 string) (*discovery.Details, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServerDetails", arg0, arg1)
	ret0, _ := ret[0].(*discovery.Details)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServerDetails indicates an expected call of GetServerDetails.
func (mr *MockDetailScannerMockRecorder) GetServerDetails(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServerDetails", reflect.TypeOf((*MockDetailScanner)(nil).GetServerDetails), arg0, arg1)
}

// MockScanner is a mock of Scanner interface.
type MockScanner struct {
	ctrl     *gomock.Controller
	recorder *MockScannerMockRecorder
}

// MockScannerMockRecorder is the mock recorder for MockScanner.
type MockScannerMockRecorder struct {
	mock *MockScanner
}

// NewMockScanner creates a new mock instance.
func NewMockScanner(ctrl *gomock.Controller) *MockScanner {
	mock := &MockScanner{ctrl: ctrl}
	mock.recorder = &MockScannerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScanner) EXPECT() *MockScannerMockRecorder {
	return m.recorder
}

// Scan mocks base method.
func (m *MockScanner) Scan() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Scan")
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *MockScannerMockRecorder) Scan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*MockScanner)(nil).Scan))
}

// Stop mocks base method.
func (m *MockScanner) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockScannerMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockScanner)(nil).Stop))
}
