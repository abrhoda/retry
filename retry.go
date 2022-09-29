package retry

import (
	"time"
)

type retryableFunction[T any] func() (T, error)

// TODO add state enum, isclosed/exhausted, channel for interrupt
type retryContext struct {
	count     int
	lastError error
}

type retryPolicy interface {
	stop(*retryContext) bool
	delay() time.Duration
}

// TODO add onerror, onopen, onclose callbacks
// TODO move context to the policy? Maybe not.
type RetryTemplate[T any] struct {
	rp retryPolicy
	rc retryContext
}

func (rt *RetryTemplate[T]) Execute(fn retryableFunction[T]) (T, error) {
	rc := retryContext{
		count:     0,
		lastError: nil,
	}

	rt.rc = rc
	var val T
	var err error

	for !rt.rp.stop(&rt.rc) {
		val, err = fn()
		if err == nil {
			break
		}
		rt.rc.count++
		rt.rc.lastError = err
		time.Sleep(rt.rp.delay())
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
