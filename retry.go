package retry

import (
  "fmt"
  "time"
)

type retryableFunction[T any] func() (T, error)

type retryPolicy interface {
  stop() bool
  delay() int
}

type SimpleRetryPolicy struct {
  MaxAttempts int
  Interval int
  count int
}

func (srp *SimpleRetryPolicy) delay() int {
  return srp.Interval
}

func (srp *SimpleRetryPolicy) stop() bool {
  srp.count++
  fmt.Printf("count = %d\n", srp.count)
  return srp.count > srp.MaxAttempts
}

func Execute[T any](rp retryPolicy, fn retryableFunction[T]) (T, error) {
  var val T
  var err error

  for !rp.stop() {
    val, err = fn()
    if err == nil {
      break
    }
    time.Sleep(time.Duration(rp.delay()) * time.Millisecond)
  }
  return val, err
}
