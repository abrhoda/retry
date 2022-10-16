# Retry
 - [Introduction](#Introduction)
 - [Examples](#Examples)
 - [Public API](#Public-API)
 - [Features and TODO List](#Features-and-TODO-List)
 - [Warning](#Warning)
 - [License](#License)

## Introduction
`retry` is the easiest way to retry functions based on policies similar to [spring-retry](https://github.com/spring-projects/spring-retry) from the [spring framework](https://spring.io/).

The basic concept of `retry` is to retry a function until is produced a non-error return value. The function signature for a retryable function is below where the function takes no parameters and the return type is a value and an error, denoted (T, error) where T is any type. 
```go
type retryableFunction[T any] func() (T, error)
```

## Examples
The follow is a basic example of how to use `retry` in a project. 
```go
policy := SimpleRetryPolicy{
  MaxAttempts: ...,
}

resp, err := template.Execute(func () (int, error) { ... })
```

## Public API
This section discusses the feautres of `retry` and it's public API.

In order to add resilience and robustness to an application, especially in a distributed system, it is sometimes useful or even necessary to retry a failed operation in case a subsequent attempt would succeed. This could be due to a brief network outage causing a downstream service to not be reachable or a spontaneous burst of traffic overloading a database for a second. `retry` painlessly solves this issue with various configurable policies per template.

### Using `RetryTemplate`
At the heart of the library is the `RetryTemplate`. The `RetryTemplate` is a simple struct with 1 field, `RetryPolicy`, and 1 main function, `Execute`, and optionally other functions to add callbacks or a channel to cancel retrying (discussed in a later section). The following is the simplest general implementation of the `RetryTemplate` with a `SimpleRetryPolicy`.

```go
maxAttempts := 10
policy := SimpleRetryPolicy{
  MaxAttempts: maxAttempts,
}

url := "http://api.myexample.com/service/health"

func checkHealth() (int, error) {
  resp, err := http.Get(url)
  if err != nil {
	  return 0, err
	}
	defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return 0, fmt.Errorf("Healthcheck response was not 200. Got %d", resp.StatusCode)
  }

  return resp.StatusCode, nil
}

template := RetryTemplate {
  RetryPolicy: policy,
}

statusCode, error := template.Execute(checkHealth())
```
Note that `checkHealth()` is a function that fits the function signature of `functionName[T any] func() (T, error)`. This is the only requirement of the function passed to the `RetryTemplate.Execute()`

### Using `RetryPolicy`
There are currently 2 types of retry policies supported in this library, with more coming in the future (see [Features and TODO List](#Feature-and-TODO-List) for all planning future features!). The 2 types of retry policies are: `SimpleRetryPolicy` and `FixedBackoffRetryPolicy`. 

The `SimpleRetryPolicy` allows for a fixed number of retries. In the following example, calling `template.Execute(myRetryableFunc)` would cause `myRetryableFunc` to be retried **a maximim of 10 times** or until a non-error value is returned from the function.
```go
maxAttempts := 10
policy := SimpleRetryPolicy{
  MaxAttempts: maxAttempts,
}

template := RetryTemplate {
  RetryPolicy: policy,
}

template.Execute(myRetryableFunc)
```

The `FixedBackoffRetryPolicy` allows for a fixed time window between retry attemps up until a limit (default=30000 ms [or 30s]) is hit. In the following example, `myRetryableFunc` will be attempted immediately initially and then every 500ms after - either until `Limit` is reached or a non-error value is returned from the function.
```go
delay := 500 * time.Millisecond
policy := FixedBackoffPolicy{
	BackoffPeriod: delay,
	Limit:         10000 * time.Millisecond,
}

template := RetryTemplate {
  RetryPolicy: policy,
}

template.Execute(myRetryableFunc)
```
### Using `RetryTemplate` Callbacks - `onOpen`, `onError`, `onClose`
It is often useful to be able take action when something happens within the context of the function that is retrying and for this purpose, `retry` provides the `onOpen`, `onError`, and `onClose` callback functions. Their individual type signatures are below.
```go
type onOpenCallbackFunction func()
type onErrorCallbackFunction func(error)
type onCloseCallbackFunction[T any] func(T, error)
```
The `onOpen` and `onClose` are the simplest callbacks and are called exactly when you might suspect: at the very beginning and at the very end, respectively. They are set through the `SetOnOpenCallback(...)` and `SetOnCloseCallback(...)` functions on the `RetryTemplate`. The `onOpen` function takes no parameters and has no return value. This may provide a place to log some information about the function being retried or the time the retries started. `onClose` takes the result final `(T, error)` from the exhuasted or finished `RetryTemplate.Execute(...)` and allows you to use the final values of `(T, error)` before returning out of the `RetryTemplate.Execute(...)` function - maybe to log a final error message or log the finish time of the retries. In the following example, we can see how `onOpen` and `onClose` are set and used to log total execution time of the retry.
```go
maxAttempts := 10
policy := SimpleRetryPolicy{
	MaxAttempts: maxAttempts,
}

openFunc := func() {
  now = time.Now()
  log.Printf("Started at: %s", now.String())
}

closeFunc := func(i int, err error) {
  now = time.Now()
  if err == nil {
    log.Printf("Finished at: %s with value: %d", now.String(), i)
  } else {
    log.Printf("Finished at: %s with value: %s", now.String(), err)
  }
}

template := RetryTemplate {
  RetryPolicy: policy
}

template.SetOnOpenCallback(openFunc)
template.SetOnCloseCallback(closeFunc)

res, err := template.Execute(func() (int, error) {
  // pick a random number from 1..6 and return error if odd 
})
```

The `onError` callback is very similar to the `onOpen` or `onClose` callbacks and is even set through the `SetOnErrorCallback(...)` function, however it is invoked on every failed attempt of the `RetryTemplate.Execute(...)` and it's parameter is the resulting failed attempt's error. This could be a different error for consecutive failed attempts and could provide a place to log errors, for example. Below is an example of how to set and use this callback function.
```go
maxAttempts := 10
policy := SimpleRetryPolicy{
	MaxAttempts: maxAttempts,
}

openFunc := func(err error) {
  now = time.Now()
  log.Printf("Error received at: %s. Error was %s", now.String(), err)
}

template := RetryTemplate {
  RetryPolicy: policy
}

template.SetOnErrorCallback(openFunc)

res, err := template.Execute(func() (int, error) {
  // pick a random number from 1..6 and return error if odd 
})
```

### Using `SetInterruptChannel`
For long running retryable functions and policies, it is convenient to have a way to signal an interrupt of the `RetryTemplate.Execute(...)` from outside of the retry itself. For this, there is the `SetInterruptChannel(<-chan bool)` function on the `RetryTemplate`. This function allows the user to supply a channel that only receives a boolean "shouldStop" value. When sent, the `RetryTemplate`'s internal context sets it's state to closed and will stop all future execution of the retryable function. This would appear to be similar to the user as if the `RetryPolicy` was exhuasted and then the `onClose` callacbk would be invoked (with the last values of the internal state `(T, error)` of the `RetryTemplate` - likely `(nil, err)` due to unsuccessful retrying) and then return the last values of `(T, error)` to the caller. Below is an example of how `SetInterruptChannel` can be set and used properly.
```go
// delay = 10s and limit = 10m
delay := 10000 * time.Millisecond
policy := FixedBackoffPolicy{
	BackoffPeriod: delay,
	Limit:         600000 * time.Millisecond,
}

template := RetryTemplate {
  RetryPolicy: policy,
}

signal := make(chan bool, 1)
template.SetInterruptChannel(signal)

go func(ch chan bool) {
  time.Sleep(5000 * time.Millisecond)
  signal <- true
}(signal)

template.Execute(myRetryableFunc)
```
Note that the function does not call `onClose` (if set) and return last `(T, error)` until **AFTER** the delay period.

## Features and TODO List
- [x] `Simple` retry policy - retry X times
- [x] Context `onopen`, `onclose`, and `onerror` callback functions
- [x] `FixedBackoff` retry policy - retry every X milliseconds for a max of X ms (deafult=30000)
- [x] `ExponentialBackoff` retry policy - retry X milliseconds (default=100ms) initially and exponentially thereafter by multiplier Y (default=2 for 100% increase) for a max of Z ms (default=30000) 
- [X] Channel to cancel `RetryTemplate.Execute` while operating.
- [ ] Require value for `ExponentialBackoffRetryPolicy`'s `InitialInterval`, `FixedBackoffRetryPolicy`'s `BackoffPeriod`, and `SimpleRetryPolicy`'s `MaxAttempts`
- [x] Check `retryContext`'s `state` at the start of `Retrypolicy`'s `delay/1` to see if we can not delay and immediately stop for cases where the `state` was set to `closed` while the function was retrying

## Warning
_This library leverages go 1.18 generics and is compatible with versions of go that support generics._

## License
Distributed under the MIT license. See `LICENSE.txt` at project root for more information. 
