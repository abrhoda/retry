package retry

import (
	"errors"
	"testing"
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

/* THIS IS A SIMPLERETRYPOLICY TEST */
func TestSimpleRetryPolicy(t *testing.T) {
	t.Run("Sends stop boolean at maxAttempts", func(t *testing.T) {
		maxAttempts := 1
		policy := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
		}

		context := retryContext{
			count:     maxAttempts,
			lastError: nil,
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
}

func TestSimpleRetryPolicyCallbacks(t *testing.T) {
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
		template.setOnOpenCallback(openFunc)
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
		template.setOnErrorCallback(errorFunc)
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
		template.setOnCloseCallback(closeFunc)
		template.Execute(
			func() (int, error) {
				return 1, nil
			},
		)
		assertTrue(t, closed)
	})
}
func createRetryTemplate[T any](rp retryPolicy) RetryTemplate[T] {
	context := retryContext{
		count:     0,
		lastError: nil,
	}

	return RetryTemplate[T]{
		rp: rp,
		rc: context,
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
