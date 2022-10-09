# Retry 
`retry` is the easiest way to retry functions based on policies similar to [spring-retry](https://github.com/spring-projects/spring-retry) from the [spring framework](https://spring.io/).

## Features and TODO list

- [x] `Simple` retry policy - retry X times
- [x] Context `onopen`, `onclose`, and `onerror` callback functions
- [x] `FixedBackoff` retry policy - retry every X milliseconds for a max of X ms (deafult=30000)
- [ ] `ExponentialBackoff` retry policy - retry X milliseconds (default=100ms) initially and exponentially thereafter by multiplier Y (default=2 for 100% increase) for a max of Z ms (default=30000) 
- [X] Channel to cancel `retryTemplate.execute` while operating.

## Getting started


## Usage
[See documentation](#)

