package retry

import (
	"errors"
	"testing"
	"time"
)

func TestRetryExecute(t *testing.T) {
	t.Run("Returns value T", func(t *testing.T) {
		srp := SimpleRetryPolicy{
			MaxAttempts: 1,
			Interval:    0,
			count:       0,
		}
		ret_val := "desired return value"
		got, err := Execute(
			&srp,
			func() (string, error) {
				return "desired return value", nil
			},
		)
		assertEqual(t, got, ret_val)
		assertNil(t, err)
	})

	t.Run("Returns error", func(t *testing.T) {
		srp := SimpleRetryPolicy{
			MaxAttempts: 1,
			Interval:    0,
			count:       0,
		}
		_, err := Execute(
			&srp,
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)

		assertError(t, err)
	})

	t.Run("Increases count on each attempt", func(t *testing.T) {
		maxAttempts := 5
		srp := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
			Interval:    0,
			count:       0,
		}
		want_count := 3
		want_string := "want_string"

		got, _ := Execute(
			&srp,
			func() (string, error) {
				if srp.count < want_count {
					return "", errors.New("Error from `Returns error`")
				} else {
					return want_string, nil
				}
			},
		)
		assertEqual(t, srp.count, want_count)
		assertEqual(t, got, want_string)
	})

	t.Run("Stops at MaxAttempts", func(t *testing.T) {
		maxAttempts := 5
		srp := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
			Interval:    0,
			count:       0,
		}
		Execute(
			&srp,
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)
		want := maxAttempts
		assertEqual(t, srp.count, want)
	})

	t.Run("Waits for delay between attempts", func(t *testing.T) {
		maxAttempts := 5
		interval := 200 * time.Millisecond // ms
		totalTime := time.Duration(maxAttempts) * interval

		srp := SimpleRetryPolicy{
			MaxAttempts: maxAttempts,
			Interval:    interval,
			count:       0,
		}

		start := time.Now()
		Execute(
			&srp,
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)
		elapsed := time.Since(start)

		if totalTime > elapsed {
			t.Error("Elapsed should take longer than total delay.")
		}

	})
}

func assertEqual[C comparable](t testing.TB, got C, want C) {
	t.Helper()
	if got != want {
		t.Errorf("got not equal want")
	}
}

func assertError(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Error("Error is nil")
	}
}

func assertNil(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Error("Error is not nil")
	}
}
