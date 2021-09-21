package sqltracing_test

// TODO: verify that the driver functions are called

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	opentracing_go "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/simplesurance/sqltracing"
	"github.com/simplesurance/sqltracing/tracing/opentracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustNewDBDriver(t *testing.T) (*mocktracer.MockTracer, string) {
	t.Helper()

	driverName := "traced-mockdb-" + fmt.Sprint(time.Now().UnixNano())

	mockTracer := mocktracer.New()

	sql.Register(
		driverName,
		sqltracing.WrapDriver(
			&nullDriver{con: &nullCon{}},
			opentracing.NewTracer(
				opentracing.WithTracer(
					func() opentracing_go.Tracer { return mockTracer },
				),
			),
		),
	)

	return mockTracer, driverName
}

func findFinishedSpan(t *testing.T, tracer *mocktracer.MockTracer, wantedSpanName string) *mocktracer.MockSpan {
	t.Helper()

	for _, span := range tracer.FinishedSpans() {
		if span.OperationName == wantedSpanName {
			return span
		}
	}

	return nil
}

func finishedSpanExist(t *testing.T, tracer *mocktracer.MockTracer, wantedSpanName string) bool {
	t.Helper()

	return findFinishedSpan(t, tracer, wantedSpanName) != nil
}

func assertHasSpan(t *testing.T, tracer *mocktracer.MockTracer, wantedSpanOp sqltracing.SQLOp) {
	t.Helper()

	assert.Truef(t, finishedSpanExist(
		t,
		tracer,
		wantedSpanOp.String(),
	),
		"span %q was not recorded", wantedSpanOp.String())
}

func assertHasNotSpan(t *testing.T, tracer *mocktracer.MockTracer, wantedSpanOp sqltracing.SQLOp) {
	t.Helper()

	assert.Falsef(t, finishedSpanExist(
		t,
		tracer,
		wantedSpanOp.String(),
	),
		"span %q should not have been recorded but was", wantedSpanOp.String())
}

func assertIsParentSpanOp(t *testing.T, tracer *mocktracer.MockTracer, parentOp, childOp sqltracing.SQLOp) {
	t.Helper()

	assertIsParentSpan(t, tracer, parentOp.String(), childOp.String())
}

func assertIsParentSpan(t *testing.T, tracer *mocktracer.MockTracer, parentOp, childOp string) {
	t.Helper()

	parentSpan := findFinishedSpan(t, tracer, parentOp)
	if parentSpan == nil {
		t.Errorf("parent span for operation %q does not exist", parentOp)
		return
	}

	childSpan := findFinishedSpan(t, tracer, childOp)
	if childSpan == nil {
		t.Errorf("child span for operation %q does not exist", childOp)
		return
	}

	if childSpan.ParentID != parentSpan.SpanContext.SpanID {
		t.Errorf("parent for %q span has ID %d, expected to be a parent of span %q with id: %d",
			childOp, childSpan.SpanContext.SpanID, parentOp, parentSpan.SpanContext.TraceID)
	}
}

func mustNewDB(t *testing.T, driverName string) *sql.DB {
	t.Helper()

	db, err := sql.Open(driverName, "")
	require.NoError(t, err, "could not open database")
	require.NotNil(t, db, "sql.open returned nil db")

	return db
}

func TestConnect(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	_, err := db.Conn(context.Background())
	require.NoError(t, err)

	assertHasSpan(t, mockTracer, sqltracing.OpSQLConnect)
}

func TestPing(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	_ = db.Ping()
	assertHasSpan(t, mockTracer, sqltracing.OpSQLPing)
}

func TestPingContext(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	err := db.PingContext(context.Background())
	require.NoError(t, err)

	assertHasSpan(t, mockTracer, sqltracing.OpSQLPing)
}

func TestExec(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	_, err := db.Exec("")
	require.NoError(t, err)

	assertHasSpan(t, mockTracer, sqltracing.OpSQLConnExec)
}

func TestExecContext(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	_, err := db.ExecContext(context.Background(), "")
	require.NoError(t, err)

	assertHasSpan(t, mockTracer, sqltracing.OpSQLConnExec)
}

func TestQuery(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	rows, err := db.Query("")
	require.NoError(t, err)
	require.NotNil(t, rows)

	rows.Next()
	rows.Close()

	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLConnQuery, sqltracing.OpSQLRowsNext)
	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLConnQuery, sqltracing.OpSQLRowsClose)
}

func TestQueryContext(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	rows, err := db.QueryContext(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, rows)

	rows.Next()
	rows.Close()

	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLConnQuery, sqltracing.OpSQLRowsNext)
	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLConnQuery, sqltracing.OpSQLRowsClose)
}

func TestTxBeginCommit(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, tx)

	_ = tx.Commit()

	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLTxBegin, sqltracing.OpSQLTxCommit)
}

func TestTxBeginRollback(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, tx)

	_ = tx.Rollback()

	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLTxBegin, sqltracing.OpSQLTxRollback)
}

func TestPrepareContext(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	stmt, err := db.PrepareContext(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, stmt)

	_ = stmt.Close()

	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLPrepare, sqltracing.OpSQLStmtClose)
}

func TestStatementExecContext(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	stmt, err := db.PrepareContext(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, stmt)

	_, err = stmt.ExecContext(context.Background())
	require.NoError(t, err)

	_ = stmt.Close()

	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLPrepare, sqltracing.OpSQLStmtClose)
	t.Run("StmtExecIsChildSpan", func(t *testing.T) {
		assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLPrepare, sqltracing.OpSQLStmtExec)
	})
}

func TestStatementQueryContext(t *testing.T) {
	mockTracer, driverName := mustNewDBDriver(t)
	db := mustNewDB(t, driverName)

	stmt, _ := db.PrepareContext(context.Background(), "")

	rows, err := stmt.QueryContext(context.Background())
	require.NoError(t, err)
	require.NotNil(t, rows)

	_ = rows.Next()
	_ = rows.Close()

	_ = stmt.Close()

	assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLPrepare, sqltracing.OpSQLStmtClose)

	t.Run("StmtQueryIsChildSpan", func(t *testing.T) {
		assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLPrepare, sqltracing.OpSQLStmtQuery)
		assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLStmtQuery, sqltracing.OpSQLRowsNext)
		assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLStmtQuery, sqltracing.OpSQLRowsClose)
	})
}

func TestWithOpsExcluded(t *testing.T) {
	driverName := "traced-mockdb-" + fmt.Sprint(time.Now().UnixNano())
	mockTracer := mocktracer.New()

	sql.Register(
		driverName,
		sqltracing.WrapDriver(
			&nullDriver{con: &nullCon{}},
			opentracing.NewTracer(
				opentracing.WithTracer(
					func() opentracing_go.Tracer { return mockTracer },
				),
			),
			sqltracing.WithOpsExcluded(
				sqltracing.OpSQLRowsNext,
				sqltracing.OpSQLConnQuery,
			),
		),
	)
	db := mustNewDB(t, driverName)

	rows, err := db.QueryContext(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, rows)

	rows.Next()
	rows.Close()

	assertHasSpan(t, mockTracer, sqltracing.OpSQLRowsClose)
	assertHasNotSpan(t, mockTracer, sqltracing.OpSQLRowsNext)
	assertHasNotSpan(t, mockTracer, sqltracing.OpSQLConnQuery)
}

func TestWithoutTracingOrphans(t *testing.T) {
	driverName := "traced-mockdb-" + fmt.Sprint(time.Now().UnixNano())

	sql.Register(
		driverName,
		sqltracing.WrapDriver(
			&nullDriver{con: &nullCon{}},
			opentracing.NewTracer(
				opentracing.WithoutTracingOrphans(),
			),
		),
	)
	db := mustNewDB(t, driverName)
	mockTracer := mocktracer.New()

	t.Run("without-parent-span", func(t *testing.T) {
		rows, err := db.QueryContext(context.Background(), "")
		require.NoError(t, err)
		require.NotNil(t, rows)

		rows.Next()
		rows.Close()

		assertHasNotSpan(t, mockTracer, sqltracing.OpSQLRowsClose)
		assertHasNotSpan(t, mockTracer, sqltracing.OpSQLRowsNext)
		assertHasNotSpan(t, mockTracer, sqltracing.OpSQLRowsClose)
		assertHasNotSpan(t, mockTracer, sqltracing.OpSQLConnQuery)
	})

	t.Run("with-parent-span", func(t *testing.T) {
		const parentSpanName = "parent"
		span := mockTracer.StartSpan(parentSpanName)
		ctx := opentracing_go.ContextWithSpan(context.Background(), span)

		rows, err := db.QueryContext(ctx, "")
		require.NoError(t, err)
		require.NotNil(t, rows)

		rows.Next()
		rows.Close()

		span.Finish()

		assertIsParentSpan(t, mockTracer, parentSpanName, sqltracing.OpSQLConnQuery.String())
		assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLConnQuery, sqltracing.OpSQLRowsNext)
		assertIsParentSpanOp(t, mockTracer, sqltracing.OpSQLConnQuery, sqltracing.OpSQLRowsClose)
	})
}
