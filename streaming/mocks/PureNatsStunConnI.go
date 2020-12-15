// Code generated by mockery v2.4.0. DO NOT EDIT.

package mocks

import (
	stan "github.com/nats-io/stan.go"
	mock "github.com/stretchr/testify/mock"
)

// PureNatsStunConnI is an autogenerated mock type for the PureNatsStunConnI type
type PureNatsStunConnI struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *PureNatsStunConnI) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Publish provides a mock function with given fields: _a0, _a1
func (_m *PureNatsStunConnI) Publish(_a0 string, _a1 []byte) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PublishAsync provides a mock function with given fields: _a0, _a1, _a2
func (_m *PureNatsStunConnI) PublishAsync(_a0 string, _a1 []byte, _a2 stan.AckHandler) (string, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, []byte, stan.AckHandler) string); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []byte, stan.AckHandler) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueueSubscribe provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *PureNatsStunConnI) QueueSubscribe(_a0 string, _a1 string, _a2 stan.MsgHandler, _a3 ...stan.SubscriptionOption) (stan.Subscription, error) {
	_va := make([]interface{}, len(_a3))
	for _i := range _a3 {
		_va[_i] = _a3[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1, _a2)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 stan.Subscription
	if rf, ok := ret.Get(0).(func(string, string, stan.MsgHandler, ...stan.SubscriptionOption) stan.Subscription); ok {
		r0 = rf(_a0, _a1, _a2, _a3...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(stan.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, stan.MsgHandler, ...stan.SubscriptionOption) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Subscribe provides a mock function with given fields: _a0, _a1, _a2
func (_m *PureNatsStunConnI) Subscribe(_a0 string, _a1 stan.MsgHandler, _a2 ...stan.SubscriptionOption) (stan.Subscription, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 stan.Subscription
	if rf, ok := ret.Get(0).(func(string, stan.MsgHandler, ...stan.SubscriptionOption) stan.Subscription); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(stan.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, stan.MsgHandler, ...stan.SubscriptionOption) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}