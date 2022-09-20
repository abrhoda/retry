package retry

import (
	"errors"
	"testing"
)

func TestRetryExecute(t *testing.T) {
	t.Run("Returns value T", func(t *testing.T) {
		got, err := execute(
			func() (int, error) {
				return 1, nil
			},
		)

		want := 1
		assertEqual(t, got, want)
		assertNil(t, err)
	})

	t.Run("Returns error", func(t *testing.T) {
		_, err := execute(
			func() (int, error) {
				return 0, errors.New("Error from `Returns error`")
			},
		)

		assertError(t, err)
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
