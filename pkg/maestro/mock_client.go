// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/owlint/maestro-client/pkg/maestro (interfaces: Maestro)

// Package maestro is a generated GoMock package.
package maestro

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMaestro is a mock of Maestro interface.
type MockMaestro struct {
	ctrl     *gomock.Controller
	recorder *MockMaestroMockRecorder
}

// MockMaestroMockRecorder is the mock recorder for MockMaestro.
type MockMaestroMockRecorder struct {
	mock *MockMaestro
}

// NewMockMaestro creates a new mock instance.
func NewMockMaestro(ctrl *gomock.Controller) *MockMaestro {
	mock := &MockMaestro{ctrl: ctrl}
	mock.recorder = &MockMaestroMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMaestro) EXPECT() *MockMaestroMockRecorder {
	return m.recorder
}

// CompleteTask mocks base method.
func (m *MockMaestro) CompleteTask(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompleteTask", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CompleteTask indicates an expected call of CompleteTask.
func (mr *MockMaestroMockRecorder) CompleteTask(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompleteTask", reflect.TypeOf((*MockMaestro)(nil).CompleteTask), arg0, arg1)
}

// Consume mocks base method.
func (m *MockMaestro) Consume(arg0 string) (*Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Consume", arg0)
	ret0, _ := ret[0].(*Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Consume indicates an expected call of Consume.
func (mr *MockMaestroMockRecorder) Consume(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Consume", reflect.TypeOf((*MockMaestro)(nil).Consume), arg0)
}

// CreateTask mocks base method.
func (m *MockMaestro) CreateTask(arg0, arg1, arg2 string, arg3 ...CreateTaskOptions) (string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateTask", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateTask indicates an expected call of CreateTask.
func (mr *MockMaestroMockRecorder) CreateTask(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTask", reflect.TypeOf((*MockMaestro)(nil).CreateTask), varargs...)
}

// DeleteTask mocks base method.
func (m *MockMaestro) DeleteTask(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTask", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTask indicates an expected call of DeleteTask.
func (mr *MockMaestroMockRecorder) DeleteTask(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTask", reflect.TypeOf((*MockMaestro)(nil).DeleteTask), arg0)
}

// FailTask mocks base method.
func (m *MockMaestro) FailTask(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FailTask", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// FailTask indicates an expected call of FailTask.
func (mr *MockMaestroMockRecorder) FailTask(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FailTask", reflect.TypeOf((*MockMaestro)(nil).FailTask), arg0)
}

// NextInQueue mocks base method.
func (m *MockMaestro) NextInQueue(arg0 string) (*Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NextInQueue", arg0)
	ret0, _ := ret[0].(*Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NextInQueue indicates an expected call of NextInQueue.
func (mr *MockMaestroMockRecorder) NextInQueue(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextInQueue", reflect.TypeOf((*MockMaestro)(nil).NextInQueue), arg0)
}

// QueueStats mocks base method.
func (m *MockMaestro) QueueStats(arg0 string) (map[string][]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueueStats", arg0)
	ret0, _ := ret[0].(map[string][]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueueStats indicates an expected call of QueueStats.
func (mr *MockMaestroMockRecorder) QueueStats(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueueStats", reflect.TypeOf((*MockMaestro)(nil).QueueStats), arg0)
}

// TaskState mocks base method.
func (m *MockMaestro) TaskState(arg0 string) (*Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TaskState", arg0)
	ret0, _ := ret[0].(*Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TaskState indicates an expected call of TaskState.
func (mr *MockMaestroMockRecorder) TaskState(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TaskState", reflect.TypeOf((*MockMaestro)(nil).TaskState), arg0)
}
