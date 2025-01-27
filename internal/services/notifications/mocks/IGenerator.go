// Code generated by mockery v1.0.0
package mocks

import mock "github.com/stretchr/testify/mock"

// IGenerator is an autogenerated mock type for the IGenerator type
type IGenerator struct {
	mock.Mock
}

// RandomCode provides a mock function with given fields:
func (_m *IGenerator) RandomCode() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// RandomToken provides a mock function with given fields:
func (_m *IGenerator) RandomToken() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
