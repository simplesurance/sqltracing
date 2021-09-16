# sqltracing
![CI](https://github.com/simplesurance/sqltracing/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/simplesurance/sqltracing)](https://goreportcard.com/report/github.com/simplesurance/sqltracing)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/simplesurance/sqltracing)

sqltracing is a Go package for tracing database operations via an OpenTracing
tracer.
It can wrap any `driver.Driver` compatible SQL driver.

It is implemented as an interceptor for
[simplesurance/sqlmw](https://github.com/simplesurance/sqlmw).

## Documentation

[Go Package Documentation](https://pkg.go.dev/github.com/simplesurance/sqltracing)

### Example

See [example_test.go](example_test.go)

## Known Issues

- Transactions: all operations on transactions except `Commit()` and
  `Rollback()` are recorded as independent spans, instead of as child spans of
  the `BeginTx()` operation

## Credits

sqltracing and simplesurance/sqlmw are based heavily on forks and the ideas of
the following projects:

- [ngrok/sqlmw](https://github.com/ngrok/sqlmw)
- [luna-duclos/instrumentedsql](https://github.com/luna-duclos/instrumentedsql)
- [ExpansiveWorlds/instrumentedsql](https://github.com/ExpansiveWorlds/instrumentedsql)
