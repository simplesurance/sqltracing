# sqlmw
![CI](https://github.com/simplesurance/sqlmw/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/simplesurance/sqlmw)](https://goreportcard.com/report/github.com/simplesurance/sqlmw)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/simplesurance/sqlmw)

sqlmw provides an absurdly simple API that allows a caller to wrap a `database/sql` driver
with middleware.

This provides an abstraction similar to http middleware or GRPC interceptors but for the database/sql package.
This allows a caller to implement observability like tracing and logging easily. More importantly, it also enables
powerful possible behaviors like transparently modifying arguments, results or query execution strategy. This power allows programmers to implement
functionality like automatic sharding, selective tracing, automatic caching, transparent query mirroring, retries, fail-over 
in a reuseable way, and more.

## Usage

- Define a new type and embed the `sqlmw.NullInterceptor` type.
- Add a method you want to intercept from the `sqlmw.Interceptor` interface.
- Wrap the driver with your interceptor with `sqlmw.Driver` and then install it with `sql.Register`.
- Use `sql.Open` on the new driver string that was passed to register.

Here's a complete example:

```go
type sqlInterceptor struct {
        sqlmw.NullInterceptor
}

func (in *sqlInterceptor) ConnQueryContext(ctx context.Context, conn driver.QueryerContext, query string, args []driver.NamedValue) (driver.Rows, error) {
        startedAt := time.Now()

        rows, err := conn.QueryContext(ctx, query, args)
        log.Printf("executed sql query, query: %s, err: %s, duration: %s", query, err, time.Since(startedAt))

        return rows, err
}

func run(dsn string) {
        // install the wrapped driver
        sql.Register("postgres-mw", sqlmw.WrapDriver(pq.Driver{}, new(sqlInterceptor)))

        db, err := sql.Open("pq-mw", "postgres://user@localhost:5432/db")
        if err != nil {
                log.Fatalln(err)
        }

        // use db object as usual
        _, _ = db.QueryContext(context.Background(), "SELECT * FROM mytable")
}
```

You may override any subset of methods to intercept in the `Interceptor` interface (https://godoc.org/github.com/simplesurance/sqlmw#Interceptor):

```go
type Interceptor interface {
        // Connection interceptors
        ConnBeginTx(context.Context, driver.ConnBeginTx, driver.TxOptions) (driver.Tx, error)
        ConnPrepareContext(context.Context, driver.ConnPrepareContext, string) (driver.Stmt, error)
        ConnPing(context.Context, driver.Pinger) error
        ConnExecContext(context.Context, driver.ExecerContext, string, []driver.NamedValue) (driver.Result, error)
        ConnQueryContext(context.Context, driver.QueryerContext, string, []driver.NamedValue) (driver.Rows, error)

        // Connector interceptors
        ConnectorConnect(context.Context, driver.Connector) (driver.Conn, error)

        // Results interceptors
        ResultLastInsertId(driver.Result) (int64, error)
        ResultRowsAffected(driver.Result) (int64, error)

        // Rows interceptors
        RowsNext(driver.Rows, []driver.Value) error
        RowsClose(driver.Rows) error

        // Stmt interceptors
        StmtExecContext(context.Context, *Stmt, []driver.NamedValue) (driver.Result, error)
        StmtQueryContext(context.Context, *Stmt, []driver.NamedValue) (driver.Rows, error)
        StmtClose(*Stmt) error

        // Tx interceptors
        TxCommit(driver.Tx) error
        TxRollback(driver.Tx) error
}
```

Bear in mind that because you are intercepting the calls entirely, that you are responsible for passing control up to the wrapped
driver in any function that you override, like so:

```go
func (in *sqlInterceptor) ConnPing(ctx context.Context, conn driver.Pinger) error {
        return conn.Ping(ctx)
}
```

## Examples

### Logging

```go
func (in *sqlInterceptor) ConnQueryContext(ctx context.Context, conn driver.QueryerContext, query string, args []driver.NamedValue) (driver.Rows, error) {
        startedAt := time.Now()

        rows, err := conn.QueryContext(ctx, query, args)
        log.Printf("executed sql query, query: %s, err: %s, duration: %s", query, err, time.Since(startedAt))

        return rows, err
}
```

### Tracing

```go
func (in *sqlInterceptor) ConnQueryContext(ctx context.Context, conn driver.QueryerContext, query string, args []driver.NamedValue) (driver.Rows, error) {
            span := trace.FromContext(ctx).NewSpan(ctx, "ConnQueryContext")
            span.Tags["query"] = query
            defer span.Finish()
            rows, err := conn.QueryContext(ctx, args)
            if err != nil {
                    span.Error(err)
            }

            return rows, err
    }
```

### Retries

```go
func (in *sqlInterceptor) ConnQueryContext(ctx context.Context, conn driver.QueryerContext, query string, args []driver.NamedValue) (driver.Rows, error) {
        for {
                rows, err := conn.QueryContext(ctx, args)
                if err == nil {
                        return rows, nil
                }

                if err != nil && !isIdempotent(query) {
                        return nil, err
                }

                select {
                case <-ctx.Done():
                        return nil, ctx.Err()

                case <-time.After(time.Second):
                }
            }
    }
```


### Forwarding Data to method calls on Stmt, Tx, Rows

See [interceptor_wrapping_example_test.go](interceptor_wrapping_example_test.go).

## Projects based on sqlmw

- Opentracing interceptor: [simplesurance/sqltracing](https://github.com/simplesurance/sqltracing)

## Go version support

Go versions 1.13 and forward are supported.

## Project Status

This is a fork of [github.com/ngrok/sqlmw](https://github.com/ngrok/sqlmw) with
the following changes:
- `driver.Stmt` returned from `PrepareContext()` can be wrapped and the custom
  type is accessible in the methods of `Stmt`.
- `StmtExecContext` and `StmtQueryContext` do not get the query string from
  `PrepareContext()`, the query can be forwarded by wrapping the `Stmt`.
- The additional `context.Context` parameter is removed from interceptor
  methods that do not have a `context.Context` parameter in their
  `database/sql` equivalent.
- `Driver()` renamed to `WrapDriver`
- No support for Go < 1.15
- Release tags
