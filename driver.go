// Package sqltracing provides a SQL driver wrapped to record trace for db
// operations.
package sqltracing

import (
	"context"
	"database/sql/driver"
	"io"

	"github.com/ngrok/sqlmw"
)

// tracedDriver wraps an SQL driver and traces all operations via an
// opentracing tracer
type tracedDriver struct {
	excludedOps map[SQLOp]struct{}
	tracer      Tracer
}

// NewDriver returns a new driver that traces all operations and forwards them
// to the passed driver.
func NewDriver(driver driver.Driver, tracer Tracer, opts ...Opt) driver.Driver {
	tracingDriver := tracedDriver{
		excludedOps: map[SQLOp]struct{}{},
		tracer:      tracer,
	}

	for _, opt := range opts {
		opt(&tracingDriver)
	}

	return sqlmw.Driver(
		driver,
		&tracingDriver,
	)
}

func (t *tracedDriver) ConnBeginTx(ctx context.Context, con driver.ConnBeginTx, txOpts driver.TxOptions) (_ driver.Tx, err error) {
	const op = OpSQLTxBegin

	if t.opIsExcluded(op) {
		return con.BeginTx(ctx, txOpts)
	}

	span, ctx := t.tracer.StartSpan(ctx, op.String())

	tx, err := con.BeginTx(ctx, txOpts)
	if err != nil {
		spanFinishFunc(span)(err)
		return nil, err
	}

	return newTracedTx(ctx, spanFinishFunc(span), tx), nil
}

func (t *tracedDriver) ConnPrepareContext(ctx context.Context, con driver.ConnPrepareContext, query string) (_ driver.Stmt, err error) {
	const op = OpSQLPrepare

	if t.opIsExcluded(op) {
		return con.PrepareContext(ctx, query)
	}

	span, ctx := t.tracer.StartSpan(ctx, op.String())
	span.SetTag(DBStatementTagKey, query)

	stmt, err := con.PrepareContext(ctx, query)
	if err != nil {
		spanFinishFunc(span)(err)
		return nil, err
	}

	return newTracedStmt(ctx, spanFinishFunc(span), stmt), nil
}

func (t *tracedDriver) ConnPing(ctx context.Context, con driver.Pinger) (err error) {
	var deferFn func(err error)

	deferFn, ctx = t.startSpan(ctx, OpSQLPing, "")
	defer deferFn(err)

	return con.Ping(ctx)
}

func (t *tracedDriver) ConnExecContext(ctx context.Context, con driver.ExecerContext, query string, args []driver.NamedValue) (_ driver.Result, err error) {
	var deferFn func(err error)

	deferFn, ctx = t.startSpan(ctx, OpSQLConnExec, query)
	defer deferFn(err)

	return con.ExecContext(ctx, query, args)
}

func (t *tracedDriver) ConnQueryContext(ctx context.Context, con driver.QueryerContext, query string, args []driver.NamedValue) (_ driver.Rows, err error) {
	const op = OpSQLConnQuery

	if t.opIsExcluded(op) {
		rows, err := con.QueryContext(ctx, query, args)
		if err != nil {
			return nil, err
		}

		// rows are wrapped to have access to the parent span of the
		// current operation, to record spans for other rows Ops that
		// are not excluded
		return newTracedRows(ctx, func(_ error) {}, rows), nil
	}

	span, ctx := t.tracer.StartSpan(ctx, op.String())
	span.SetTag(DBStatementTagKey, query)

	rows, err := con.QueryContext(ctx, query, args)
	if err != nil {
		spanFinishFunc(span)(err)
		return nil, err
	}

	return newTracedRows(ctx, spanFinishFunc(span), rows), nil
}

func (t *tracedDriver) ConnectorConnect(ctx context.Context, connector driver.Connector) (_ driver.Conn, err error) {
	var deferFn func(err error)

	deferFn, ctx = t.startSpan(ctx, OpSQLConnect, "")
	defer deferFn(err)

	return connector.Connect(ctx)
}

func (t *tracedDriver) ResultLastInsertId(res driver.Result) (int64, error) {
	return res.LastInsertId()
}

func (t *tracedDriver) ResultRowsAffected(res driver.Result) (int64, error) {
	return res.RowsAffected()
}

func (t *tracedDriver) RowsNext(ctx context.Context, rows driver.Rows, dest []driver.Value) (err error) {
	tracedCtx := ctx

	if tracedRows, ok := rows.(*tracedRows); ok {
		tracedCtx = tracedRows.ctx
	}

	deferFn, _ := t.startSpan(tracedCtx, OpSQLRowsNext, "", io.EOF)
	defer deferFn(err)

	return rows.Next(dest)
}

func (t *tracedDriver) RowsClose(ctx context.Context, rows driver.Rows) (err error) {
	const op = OpSQLRowsClose

	if tracedRows, ok := rows.(*tracedRows); ok {
		deferFn, _ := t.startSpan(tracedRows.ctx, op, "")
		defer deferFn(err)

		// nil instead of err is passed because it finishes the operation that
		// created the Stmt, which succeeded
		defer tracedRows.parentSpanFinishFn(nil)

		return rows.Close()
	}

	deferFn, _ := t.startSpan(ctx, op, "")
	defer deferFn(err)

	return rows.Close()
}

func (t *tracedDriver) StmtExecContext(ctx context.Context, stmt driver.StmtExecContext, _ string, args []driver.NamedValue) (_ driver.Result, err error) {
	// TODO: record as child span of PrepareContext span
	deferFn, ctx := t.startSpan(ctx, OpSQLStmtExec, "")
	defer deferFn(err)

	return stmt.ExecContext(ctx, args)
}

func (t *tracedDriver) StmtQueryContext(ctx context.Context, stmt driver.StmtQueryContext, _ string, args []driver.NamedValue) (rows driver.Rows, err error) {
	// TODO: record as child span of PrepareContext span

	deferFn, ctx := t.startSpan(ctx, OpSQLStmtQuery, "")
	defer deferFn(err)

	return stmt.QueryContext(ctx, args)
}

func (t *tracedDriver) StmtClose(ctx context.Context, stmt driver.Stmt) (err error) {
	if tracedStmt, ok := stmt.(*tracedStmt); ok {
		deferFn, _ := t.startSpan(tracedStmt.ctx, OpSQLStmtClose, "")
		defer deferFn(err)

		// nil instead of err is passed because it finishes the operation that
		// created the Stmt, which succeeded
		defer tracedStmt.parentSpanFinishFn(nil)

		return stmt.Close()
	}

	deferFn, _ := t.startSpan(ctx, OpSQLStmtClose, "")
	defer deferFn(nil)

	return stmt.Close()
}

func (t *tracedDriver) TxCommit(ctx context.Context, tx driver.Tx) (err error) {
	const op = OpSQLTxCommit

	if tracedTx, ok := tx.(*tracedTx); ok {
		deferFn, _ := t.startSpan(tracedTx.ctx, op, "")
		defer deferFn(err)
		tracedTx.parentSpanFinishFn(nil)

		return tx.Commit()
	}

	deferFn, _ := t.startSpan(ctx, op, "")
	defer deferFn(err)

	return tx.Commit()
}

func (t *tracedDriver) TxRollback(ctx context.Context, tx driver.Tx) (err error) {
	const op = OpSQLTxRollback

	if tracedTx, ok := tx.(*tracedTx); ok {
		deferFn, _ := t.startSpan(tracedTx.ctx, op, "")
		defer deferFn(err)
		tracedTx.parentSpanFinishFn(nil)

		return tx.Rollback()
	}

	deferFn, _ := t.startSpan(ctx, op, "")
	defer deferFn(err)

	return tx.Rollback()
}

func (t *tracedDriver) opIsExcluded(op SQLOp) bool {
	_, exist := t.excludedOps[op]
	return exist
}

var _ sqlmw.Interceptor = &tracedDriver{}
