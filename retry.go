package retry

import (
	"time"
)

type retryableFunction[T any] func() (T, error)

type retryPolicy interface {
	stop() bool
	delay() time.Duration
}

type SimpleRetryPolicy struct {
	MaxAttempts int
	Interval    time.Duration
	count       int
}

func (srp *SimpleRetryPolicy) delay() time.Duration {
	return srp.Interval
}

func (srp *SimpleRetryPolicy) stop() bool {
	srp.count++
	return srp.count >= srp.MaxAttempts
}

func Execute[T any](rp retryPolicy, fn retryableFunction[T]) (T, error) {
	var val T
	var err error
	stop := false
	for !stop {
		stop = rp.stop()
		val, err = fn()
		if err == nil {
			break
		}
		time.Sleep(rp.delay())
	}
	return val, err
}
