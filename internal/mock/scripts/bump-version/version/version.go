// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/robgonnella/ops/internal/scripts/bump-version/version (interfaces: VersionControl,VersionGenerator)
//
// Generated by this command:
//
//	mockgen -destination=../../../mock/scripts/bump-version/version/version.go -package=mock_version . VersionControl,VersionGenerator
//

// Package mock_version is a generated GoMock package.
package mock_version

import (
	reflect "reflect"

	version "github.com/robgonnella/ops/internal/scripts/bump-version/version"
	gomock "go.uber.org/mock/gomock"
)

// MockVersionControl is a mock of VersionControl interface.
type MockVersionControl struct {
	ctrl     *gomock.Controller
	recorder *MockVersionControlMockRecorder
}

// MockVersionControlMockRecorder is the mock recorder for MockVersionControl.
type MockVersionControlMockRecorder struct {
	mock *MockVersionControl
}

// NewMockVersionControl creates a new mock instance.
func NewMockVersionControl(ctrl *gomock.Controller) *MockVersionControl {
	mock := &MockVersionControl{ctrl: ctrl}
	mock.recorder = &MockVersionControlMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVersionControl) EXPECT() *MockVersionControlMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockVersionControl) Add(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockVersionControlMockRecorder) Add(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockVersionControl)(nil).Add), arg0)
}

// Commit mocks base method.
func (m *MockVersionControl) Commit(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit.
func (mr *MockVersionControlMockRecorder) Commit(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockVersionControl)(nil).Commit), arg0)
}

// Tag mocks base method.
func (m *MockVersionControl) Tag(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tag", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Tag indicates an expected call of Tag.
func (mr *MockVersionControlMockRecorder) Tag(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tag", reflect.TypeOf((*MockVersionControl)(nil).Tag), arg0)
}

// MockVersionGenerator is a mock of VersionGenerator interface.
type MockVersionGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockVersionGeneratorMockRecorder
}

// MockVersionGeneratorMockRecorder is the mock recorder for MockVersionGenerator.
type MockVersionGeneratorMockRecorder struct {
	mock *MockVersionGenerator
}

// NewMockVersionGenerator creates a new mock instance.
func NewMockVersionGenerator(ctrl *gomock.Controller) *MockVersionGenerator {
	mock := &MockVersionGenerator{ctrl: ctrl}
	mock.recorder = &MockVersionGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVersionGenerator) EXPECT() *MockVersionGeneratorMockRecorder {
	return m.recorder
}

// Generate mocks base method.
func (m *MockVersionGenerator) Generate(arg0 version.VersionData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Generate indicates an expected call of Generate.
func (mr *MockVersionGeneratorMockRecorder) Generate(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockVersionGenerator)(nil).Generate), arg0)
}
