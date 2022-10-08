package retry

import (
	"time"
)

type retryableFunction[T any] func() (T, error)
type onOpenCallbackFunction func()
type onErrorCallbackFunction func(error)
type onCloseCallbackFunction[T any] func(T, error)

// TODO add state enum, isclosed/exhausted, channel for interrupt
type retryContext struct {
	count     int
	lastError error
}

type retryPolicy interface {
	stop(*retryContext) bool
	delay() time.Duration
}

type RetryTemplate[T any] struct {
	rp      retryPolicy
	rc      retryContext
	onOpen  onOpenCallbackFunction
	onClose onCloseCallbackFunction[T]
	onError onErrorCallbackFunction
}

func (rt *RetryTemplate[T]) setOnOpenCallback(fn onOpenCallbackFunction) {
	rt.onOpen = fn
}

func (rt *RetryTemplate[T]) setOnCloseCallback(fn onCloseCallbackFunction[T]) {
	rt.onClose = fn
}

func (rt *RetryTemplate[T]) setOnErrorCallback(fn onErrorCallbackFunction) {
	rt.onError = fn
}

func (rt *RetryTemplate[T]) Execute(fn retryableFunction[T]) (T, error) {
	rc := retryContext{
		count:     0,
		lastError: nil,
	}

	rt.rc = rc
	var val T
	var err error

	if rt.onOpen != nil {
		rt.onOpen()
	}

	for !rt.rp.stop(&rt.rc) {
		rt.rc.count++

		val, err = fn()
		if err == nil {
			break
		}

		rt.rc.lastError = err
		if rt.onError != nil {
			rt.onError(err)
		}

		time.Sleep(rt.rp.delay())
	}

	if rt.onClose != nil {
		rt.onClose(val, err)
	}

	return val, err
}

/*
Simple Retry Policy impl
*/
type SimpleRetryPolicy struct {
	MaxAttempts int
}

func (srp SimpleRetryPolicy) delay() time.Duration {
	return 0
}

func (srp SimpleRetryPolicy) stop(rc *retryContext) bool {
	return rc.count >= srp.MaxAttempts
}

/*
Fixed Backoff Policy impl
*/
type FixedBackoffPolicy struct {
	BackoffPeriod time.Duration
	Limit         time.Duration
}

func (fbp FixedBackoffPolicy) delay() time.Duration {
	return fbp.BackoffPeriod
}

func (fbp FixedBackoffPolicy) stop(rc *retryContext) bool {
	// use a default limit if one is not provided
	if fbp.Limit == 0 {
		fbp.Limit = 30000
	}

	return (fbp.BackoffPeriod * time.Duration(rc.count)) >= fbp.Limit
}
