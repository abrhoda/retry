package retry

import (
	"errors"
	"testing"
)

func TestRetryExecute(t *testing.T) {
	t.Run("Returns value T", func(t *testing.T) {
    srp := SimpleRetryPolicy{
      MaxAttempts: 1,
      Interval: 0,
      count: 0,
    }
    ret_val := "desired return value"
		got, err := Execute(
      &srp,
			func() (string, error) {
				return "desired return value", nil
			},
		)

    assertEqual(t, srp.count, 1)
		assertEqual(t, got, ret_val)
		assertNil(t, err)
	})

	t.Run("Returns error", func(t *testing.T) {
    srp := SimpleRetryPolicy{
      MaxAttempts: 1,
      Interval: 0,
      count: 0,
    }
		_, err := Execute(
      &srp,
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)

		assertError(t, err)
	})

  t.Run("Stops at MaxAttempts", func(t *testing.T) {
    maxAttempts := 5
    srp := SimpleRetryPolicy{
      MaxAttempts: maxAttempts,
      Interval: 0,
      count: 0,
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
