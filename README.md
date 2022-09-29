# Retry 
`retry` is the easiest way to retry functions based on policies similar to [spring-retry](https://github.com/spring-projects/spring-retry) from the [spring framework](https://spring.io/).

## Features and TODO list

- [x] `Simple` retry policy - retry X times
- [ ] Context `onopen`, `onclose`, and `onerror` callback functions
- [ ] `FixedBackoff` retry policy - retry every X milliseconds for a max of X ms (deafult=30000)
- [ ] `ExponentialBackoff` retry policy - retry X milliseconds (default=100ms) initially and exponentially thereafter by multiplier Y (default=2 for100% increase) for a max of Z ms (default=30000) 

## Getting started


## Usage
[See documentation](#)

