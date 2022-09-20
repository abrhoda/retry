package retry

type retryableFunction[T any] func() (T, error)

func execute[T any](fn retryableFunction[T]) (T, error) {
	return fn()
}
