package retry

import (
	"time"
)

type retryableFunction[T any] func() (T, error)

type retryPolicy interface {
	stop(int) bool
	delay() time.Duration
}

type SimpleRetryPolicy struct {
	MaxAttempts int
	Interval    time.Duration
}

func (srp *SimpleRetryPolicy) delay() time.Duration {
	return srp.Interval
}

func (srp *SimpleRetryPolicy) stop(count int) bool {
	return count >= srp.MaxAttempts
}

type RetryTemplate[T any] struct {
  retryPolicy retryPolicy
  count int
}

/*
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
*/

func (rt *RetryTemplate[T]) execute(fn retryableFunction[T]) (T, error) {
  var val T
  var err error
  
  stop := false
	for !stop {
		stop = rt.retryPolicy.stop(rt.count)
		val, err = fn()
		if err == nil {
			break
		}
		time.Sleep(rt.retryPolicy.delay())
	}

  return val, err
}
