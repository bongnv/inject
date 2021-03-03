# inject

[![Go Reference](https://pkg.go.dev/badge/github.com/bongnv/inject.svg)](https://pkg.go.dev/github.com/bongnv/inject)
[![Build](https://github.com/bongnv/inject/workflows/Build/badge.svg)](https://github.com/bongnv/inject/actions?query=workflow%3ABuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/bongnv/inject)](https://goreportcard.com/report/github.com/bongnv/inject)
[![codecov](https://codecov.io/gh/bongnv/inject/branch/main/graph/badge.svg?token=RP3ua8huXh)](https://codecov.io/gh/bongnv/inject)

A dependency injection library for Go that supports:

- Registering and injecting dependencies by names
- Registering dependencies by factory functions

## Installation

```
go get github.com/bongnv/inject
```

## Usages

The following example demonstrates the use of factory functions for registering dependencies:

```go
func loadAppConfig() *AppConfig {
  // load app config here
}

func newLogger(cfg *AppConfig) (Logger, error) {
  // initialize a logger
  return &loggerImpl{}
}

// ServiceA has Logger as a dependency
type ServiceA struct {
  Logger Logger `inject:"logger"`
}

func newServiceA() (*ServiceA, error) {
  // init your serviceA here
}

// init func
func initDependencies() {
  c := inject.New()
  c.MustRegister("config", loadAppConfig)
  c.MustRegister("logger", newLogger), 
  // serviceA will created and registered, logger will also be injected
  c.MustRegister("serviceA", newServiceA),
}
```
