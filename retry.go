package retry

import (
	"sync"
	"time"
  "math"
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
	delay(*retryContext) time.Duration
}

type RetryTemplate[T any] struct {
	RetryPolicy retryPolicy
	rc          *retryContext
	onOpen      onOpenCallbackFunction
	onClose     onCloseCallbackFunction[T]
	onError     onErrorCallbackFunction
	recv        <-chan bool
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

	for !rt.RetryPolicy.stop(rt.rc) {
		rt.rc.count++
		val, err = fn()
		if err == nil {
			break
		}

		rt.rc.lastError = err

		if rt.onError != nil {
			rt.onError(err)
		}
    delay := rt.RetryPolicy.delay(rt.rc)
		time.Sleep(delay)
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

func (srp SimpleRetryPolicy) delay(rc *retryContext) time.Duration {
	return 0
}

func (srp SimpleRetryPolicy) stop(rc *retryContext) bool {
  // TODO require MaxAttempts

	if isContextClosed(rc) {
		return true
	}

	return rc.count >= srp.MaxAttempts
}

/*
Fixed Backoff Policy impl
*/
type FixedBackoffRetryPolicy struct {
	BackoffPeriod time.Duration
	Limit         time.Duration
}

func (fbp *FixedBackoffRetryPolicy) delay(rc *retryContext) time.Duration {
  if isContextClosed(rc) {
    return 0
  }

	return fbp.BackoffPeriod
}

func (fbp *FixedBackoffRetryPolicy) stop(rc *retryContext) bool {
  // TODO require Backoff period or provide default

	if isContextClosed(rc) {
		return true
	}

	// use a default limit if one is not provided
	if fbp.Limit == 0 {
		fbp.Limit = 30000 * time.Millisecond
	}

	return (fbp.BackoffPeriod * time.Duration(rc.count)) >= fbp.Limit
}

/*
Exponential Backoff Policy impl
*/
type ExponentialBackoffRetryPolicy struct {
	InitialInterval time.Duration
	Multiplier      float64
	Limit           time.Duration
}

func (ebp ExponentialBackoffRetryPolicy) delay(rc *retryContext) time.Duration {
  if isContextClosed(rc) {
    return 0
  }

  if ebp.Multiplier == 0 {
    ebp.Multiplier = 2
  }

  if ebp.Limit == 0 {
    ebp.Limit = 30000 * time.Millisecond
  }

  // TODO require an InitialInterval
  next := ebp.InitialInterval * time.Duration(math.Pow(ebp.Multiplier, float64(rc.count-1)))

  if next > ebp.Limit {
    next = ebp.Limit
  }
  return next
}

func (ebp ExponentialBackoffRetryPolicy) stop(rc *retryContext) bool {
	return isContextClosed(rc)
}

func isContextClosed(rc *retryContext) bool {
	rc.mu.Lock()
  defer rc.mu.Unlock()
  return rc.state == closed
}
