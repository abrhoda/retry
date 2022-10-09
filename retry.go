package retry

import (
	"sync"
	"time"
)

type retryableFunction[T any] func() (T, error)
type onOpenCallbackFunction func()
type onErrorCallbackFunction func(error)
type onCloseCallbackFunction[T any] func(T, error)

type retryContextState int

const (
	opened retryContextState = iota
	closed
)

type retryContext struct {
	count     int
	lastError error
	state     retryContextState
	mu        sync.Mutex
}

type retryPolicy interface {
	stop(*retryContext) bool
	delay() time.Duration
}

type RetryTemplate[T any] struct {
	rp      retryPolicy
	rc      *retryContext
	onOpen  onOpenCallbackFunction
	onClose onCloseCallbackFunction[T]
	onError onErrorCallbackFunction
	recv    <-chan bool
}

func (rt *RetryTemplate[T]) SetOnOpenCallback(fn onOpenCallbackFunction) {
	rt.onOpen = fn
}

func (rt *RetryTemplate[T]) SetOnCloseCallback(fn onCloseCallbackFunction[T]) {
	rt.onClose = fn
}

func (rt *RetryTemplate[T]) SetOnErrorCallback(fn onErrorCallbackFunction) {
	rt.onError = fn
}

func (rt *RetryTemplate[T]) SetInterruptChannel(recv <-chan bool) {
	rt.recv = recv
}

func (rt *RetryTemplate[T]) Execute(fn retryableFunction[T]) (T, error) {
	rc := retryContext{
		count:     0,
		lastError: nil,
		state:     opened,
	}

	rt.rc = &rc
	var val T
	var err error

	if rt.onOpen != nil {
		rt.onOpen()
	}

	if rt.recv != nil {
		go func(recv <-chan bool) {
			<-recv
			rt.rc.mu.Lock()
			defer rt.rc.mu.Unlock()
			rt.rc.state = closed
		}(rt.recv)
	}

	for !rt.rp.stop(rt.rc) {
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
	// check retryContextState
	rc.mu.Lock()
	current := rc.state
	rc.mu.Unlock()

	if current == closed {
		return true
	}

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
	// check retryContextState
	rc.mu.Lock()
	current := rc.state
	rc.mu.Unlock()

	if current == closed {
		return true
	}

	// use a default limit if one is not provided
	if fbp.Limit == 0 {
		fbp.Limit = 30000
	}

	return (fbp.BackoffPeriod * time.Duration(rc.count)) >= fbp.Limit
}
