package sqltracing_test

import (
	"context"
	"database/sql"

	"github.com/simplesurance/sqltracing"
	"github.com/simplesurance/sqltracing/tracing/opentracing"
)

func ExampleWrapDriver() {
	// register an sql driver called "traced-sql", that wraps the passed
	// driver.
	// Instead of nullDriver, pgx.GetDefaultDriver() can be passed for
	// example to trace operations of the pgx driver.
	sql.Register(
		"traced-sql",
		sqltracing.WrapDriver(&nullDriver{}, opentracing.NewTracer()),
	)

	// open a database to the passed dsn, using the traced-sql driver
	db, _ := sql.Open("traced-sql", "postgres://localhost:5432")

	// use the db as usual, operations that have context parameter should
	// be used to be able to create child spans for parent operations,
	// instead of new root tracing spans for each db operation
	_, _ = db.QueryContext(context.Background(), "SELECT * FROM mytable")
}
