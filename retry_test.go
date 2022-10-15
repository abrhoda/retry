package retry

import (
	"errors"
	"testing"
	"time"
  //"fmt"
)

func TestRetryTemplate(t *testing.T) {
	t.Run("Execute returns value T", func(t *testing.T) {
		maxAttempts := 1
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)
		want := 1
		got, err := template.Execute(
			func() (int, error) {
				return 1, nil
			},
		)
		assertEqual(t, got, want)
		assertErrorNil(t, err)
	})

	t.Run("Execute returns error", func(t *testing.T) {
		maxAttempts := 1
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)
		_, err := template.Execute(
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)

		assertErrorNotNil(t, err)
	})

	t.Run("Recv sets context state to closed on receiving signal", func(t *testing.T) {
		signal := make(chan bool, 1)
		maxAttempts := 100
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)
		template.SetInterruptChannel(signal)

		go func(ch chan bool) {
			time.Sleep(250 * time.Millisecond)
			signal <- true
		}(signal)

    _, err := template.Execute(
			func() (int, error) {
				time.Sleep(100 * time.Millisecond)
				return 0, errors.New("")
			},
		)

		assertEqual(t, template.rc.state, closed)
		assertNotEqual(t, template.rc.count, maxAttempts)
    assertErrorNotNil(t, err)
	})
}

func TestRetryContext(t *testing.T) {
	t.Run("Increases count on each retry attempt", func(t *testing.T) {
		maxAttempts := 5
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)

		want_count := 3

		template.Execute(
			func() (int, error) {
				if template.rc.count < want_count {
					return 0, errors.New("Error from `Returns error`")
				} else {
					return 1, nil
				}
			},
		)
		assertEqual(t, template.rc.count, want_count)
	})

	t.Run("Sets lastError on the context on failure attempt", func(t *testing.T) {
		maxAttempts := 1
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)

		template.Execute(
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)

		assertErrorNotNil(t, template.rc.lastError)
	})
}

func TestIsContextClosed(t *testing.T) {
  t.Run("Returns the current retryContext.state", func(t *testing.T) {
    context := retryContext{}
    isOpened := isContextClosed(&context)
    context.state = closed
    isClosed := isContextClosed(&context)

    assertFalse(t, isOpened)
    assertTrue(t, isClosed)
  })
}

func TestSimpleRetryPolicy(t *testing.T) {
	t.Run("Sends stop boolean at maxAttempts", func(t *testing.T) {
		maxAttempts := 1
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		context := retryContext{
			count:     maxAttempts,
			lastError: nil,
			state:     opened,
		}

		stop := policy.stop(&context)
		assertTrue(t, stop)
	})

	t.Run("Execute stops at MaxAttempts", func(t *testing.T) {
		maxAttempts := 5
		srp := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](srp)
		template.Execute(
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)

		want := maxAttempts
		assertEqual(t, template.rc.count, want)
	})

	t.Run("Stops when retryContext state is closed", func(t *testing.T) {
		maxAttempts := 1
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		context := retryContext{
			count:     maxAttempts,
			lastError: nil,
			state:     opened,
		}

		context.state = closed
		stop := policy.stop(&context)
		assertTrue(t, stop)
	})
}

func TestFixedBackoffRetryPolicy(t *testing.T) {
	t.Run("Waits `BackoffPeriod` miliseconds between reties", func(t *testing.T) {
		delay := 1000 * time.Millisecond
		policy := FixedBackoffRetryPolicy{
			BackoffPeriod: delay,
			Limit:         30000 * time.Millisecond,
		}

		context := retryContext{
			count:     0,
			lastError: nil,
			state:     opened,
		}
		assertEqual(t, policy.delay(&context), delay)
	})

	t.Run("Retries until Limit is reached", func(t *testing.T) {
		delay := 1000 * time.Millisecond
		policy := FixedBackoffRetryPolicy{
			BackoffPeriod: delay,
			Limit:         5000 * time.Millisecond,
		}

		context := retryContext{
			count:     0,
			lastError: nil,
			state:     opened,
		}

		stop := policy.stop(&context)
		assertFalse(t, stop)

		context.count = 5

		stop = policy.stop(&context)
		assertTrue(t, stop)
	})
	t.Run("Uses a default `Limit` if none is set", func(t *testing.T) {
		delay := 1000 * time.Millisecond
		policy := FixedBackoffRetryPolicy{
			BackoffPeriod: delay,
		}

		context := retryContext{
			count:     0,
			lastError: nil,
			state:     opened,
		}

		context.count = 30

		stop := policy.stop(&context)
		assertTrue(t, stop)
	})

	t.Run("Stops when retryContext state is closed", func(t *testing.T) {
		delay := 1000 * time.Millisecond
		policy := FixedBackoffRetryPolicy{
			BackoffPeriod: delay,
		}

		context := retryContext{
			count:     0,
			lastError: nil,
			state:     opened,
		}

		context.state = closed
		stop := policy.stop(&context)
		assertTrue(t, stop)
	})

  t.Run("Delay no time when retryContextState is closed", func(t *testing.T) {
		delay := 1000 * time.Millisecond
		policy := FixedBackoffRetryPolicy{
			BackoffPeriod: delay,
		}

		context := retryContext{
			count:     0,
			lastError: nil,
			state:     opened,
		}

    context.count++

    hasDelay := (0 < policy.delay(&context))
    context.state = closed
    hasNoDelay := (0 == policy.delay(&context))

    
    assertTrue(t, hasDelay)
    assertTrue(t, hasNoDelay)

  })
}
func TestExponentialBackoffRetryPolicy(t *testing.T) {
	t.Run("Delay has a default multiplier and limit if none are set", func(t *testing.T) {
    interval := 500 * time.Millisecond
    
    policy := ExponentialBackoffRetryPolicy {
      InitialInterval: interval,  
    }

		context := retryContext{
			count:     0,
			lastError: nil,
			state:     opened,
		}

    context.count++
    first := policy.delay(&context)
    context.count++
    second := policy.delay(&context)
    context.count++
    third := policy.delay(&context)
    context.count++
    fourth := policy.delay(&context)

    context.count++ //8000
    context.count++ //16000
    context.count++ //32000
    passedLimit := policy.delay(&context)

    // calculated off interval*(multiplier^count) = 500, 1000, 2000, 4000
    assertEqual(t, first, 500*time.Millisecond)
    assertEqual(t, second, 1000*time.Millisecond)
    assertEqual(t, third, 2000*time.Millisecond)
    assertEqual(t, fourth, 4000*time.Millisecond)
    assertEqual(t, passedLimit, 30000*time.Millisecond)
  })

	t.Run("Delay back off time is calculated as a function of (base*(multiplier^n))", func(t *testing.T) {
    interval := 300 * time.Millisecond
    multiplier := 3.0
    
    policy := ExponentialBackoffRetryPolicy {
      InitialInterval: interval,  
      Multiplier: multiplier,
    }

		context := retryContext{
			count:     0,
			lastError: nil,
			state:     opened,
		}

    context.count++
    first := policy.delay(&context)
    context.count++
    second := policy.delay(&context)
    context.count++
    third := policy.delay(&context)
    context.count++
    fourth := policy.delay(&context)

    assertEqual(t, first, 300*time.Millisecond)
    assertEqual(t, second, 900*time.Millisecond)
    assertEqual(t, third, 2700*time.Millisecond)
    assertEqual(t, fourth, 8100*time.Millisecond)

  })

  t.Run("Stop only when retryContextState is closed", func(t *testing.T) {
    interval := 500 * time.Millisecond
    
    policy := ExponentialBackoffRetryPolicy {
      InitialInterval: interval,  
    }

		context := retryContext{
			state:     opened,
		}
    
    openedShouldStop := policy.stop(&context)
    context.state = closed
    closedShouldStop := policy.stop(&context)

    assertFalse(t, openedShouldStop)
    assertTrue(t, closedShouldStop)

  })

  t.Run("Delay no time when retryContextState is closed", func(t *testing.T) {
    interval := 500 * time.Millisecond
    
    policy := ExponentialBackoffRetryPolicy {
      InitialInterval: interval,  
    }

		context := retryContext{
			state:     opened,
		}

    context.count++

    hasDelay := (0 < policy.delay(&context))
    context.state = closed
    hasNoDelay := (0 == policy.delay(&context))

    
    assertTrue(t, hasDelay)
    assertTrue(t, hasNoDelay)

  })
}
func TestRetryTemplateCallbacks(t *testing.T) {
	t.Run("Pass through when no callbacks are set", func(t *testing.T) {
		maxAttempts := 5
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)

		assertCallbackNil[int](t, template.onOpen)
		assertCallbackNil[int](t, template.onError)
		assertCallbackNil[int](t, template.onClose)
	})

	t.Run("Calls onOpen when it's set at the beginning only", func(t *testing.T) {
		var opened bool

		openFunc := func() {
			opened = true
		}

		maxAttempts := 1
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)
		template.SetOnOpenCallback(openFunc)
		template.Execute(
			func() (int, error) {
				return 1, nil
			},
		)
		assertTrue(t, opened)

	})

	t.Run("Calls onError when it's set on every failed attempt", func(t *testing.T) {
		var total int

		errorFunc := func(e error) {
			total++
		}

		maxAttempts := 5
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)
		template.SetOnErrorCallback(errorFunc)
		template.Execute(
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)
		assertEqual(t, total, maxAttempts)
	})

	t.Run("Calls onClose when it's set at the end only", func(t *testing.T) {
		var closed bool

		closeFunc := func(i int, e error) {
			closed = true
		}

		maxAttempts := 5
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		template := createRetryTemplate[int](policy)
		template.SetOnCloseCallback(closeFunc)
		template.Execute(
			func() (int, error) {
				return 1, nil
			},
		)
		assertTrue(t, closed)
	})
}

/* HELPER FUNCTIONS */
func createRetryTemplate[T any](rp retryPolicy) RetryTemplate[T] {
	return RetryTemplate[T]{
		RetryPolicy:   rp,
		rc:   nil,
		recv: nil,
	}
}

func assertTrue(t testing.TB, check bool) {
	t.Helper()
	if !check {
		t.Errorf("check is false")
	}
}

func assertFalse(t testing.TB, check bool) {
	t.Helper()
	if check {
		t.Errorf("check is true")
	}
}
func assertEqual[C comparable](t testing.TB, got C, want C) {
	t.Helper()
	if got != want {
		t.Errorf("got not equal want")
	}
}

func assertNotEqual[C comparable](t testing.TB, got C, want C) {
	t.Helper()
	if got == want {
		t.Errorf("got equal want")
	}
}

func assertErrorNotNil(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Error("error is nil")
	}
}

func assertErrorNil(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Error("error is not nil")
	}
}

type Callback[T any] interface {
	onOpenCallbackFunction | onErrorCallbackFunction | onCloseCallbackFunction[T]
}

func assertCallbackNil[T any, C Callback[T]](t testing.TB, fn C) {
	t.Helper()
	if fn != nil {
		t.Error("Callback was not nil")
	}
}

func assertCallbackNotNil[T any, C Callback[T]](t testing.TB, fn C) {
	t.Helper()
	if fn == nil {
		t.Error("Callback was nil")
	}
}
