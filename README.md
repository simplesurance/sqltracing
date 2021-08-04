# sqltracing

sqltracing is a Go package for tracing database operations via an OpenTracing
tracer.
It can wrap any `driver.Driver` compatible SQL driver.

## Example

See [example_test.go](example_test.go)

## Known Issues

- Prepared Statements: `ExecContext()` and `QueryContext()` operations are
  recorded as independent spans instead of as child spawns of the
  `PrepareContext()` operation.
- Transactions: all operations on transactions except `Commit()` and
  `Rollback()` are recorded as independent spans, instead of as child spans of
  the `BeginTx()` operation

## History

sqltracing is build on top of the [ngrok/sqlmw](https://github.com/ngrok/sqlmw)
SQL middleware and contains some modified code of
[luna-duclos/instrumentedsql](https://github.com/luna-duclos/instrumentedsql).
\
[ngrok/sqlmw](https://github.com/ngrok/sqlmw) is also based on
[luna-duclos/instrumentedsql](https://github.com/luna-duclos/instrumentedsql)
which is a fork of
[ExpansiveWorlds(instrumentedsql](https://github.com/ExpansiveWorlds/instrumentedsql).
