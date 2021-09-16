package sqlmw

import (
	"context"
	"database/sql/driver"
)

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

var _ Interceptor = NullInterceptor{}

// NullInterceptor is a complete passthrough interceptor that implements every method of the Interceptor
// interface and performs no additional logic. Users should Embed it in their own interceptor so that they
// only need to define the specific functions they are interested in intercepting.
type NullInterceptor struct{}

func (NullInterceptor) ConnBeginTx(ctx context.Context, conn driver.ConnBeginTx, txOpts driver.TxOptions) (driver.Tx, error) {
	return conn.BeginTx(ctx, txOpts)
}

func (NullInterceptor) ConnPrepareContext(ctx context.Context, conn driver.ConnPrepareContext, query string) (driver.Stmt, error) {
	return conn.PrepareContext(ctx, query)
}

func (NullInterceptor) ConnPing(ctx context.Context, conn driver.Pinger) error {
	return conn.Ping(ctx)
}

func (NullInterceptor) ConnExecContext(ctx context.Context, conn driver.ExecerContext, query string, args []driver.NamedValue) (driver.Result, error) {
	return conn.ExecContext(ctx, query, args)
}

func (NullInterceptor) ConnQueryContext(ctx context.Context, conn driver.QueryerContext, query string, args []driver.NamedValue) (driver.Rows, error) {
	return conn.QueryContext(ctx, query, args)
}

func (NullInterceptor) ConnectorConnect(ctx context.Context, connect driver.Connector) (driver.Conn, error) {
	return connect.Connect(ctx)
}

func (NullInterceptor) ResultLastInsertId(res driver.Result) (int64, error) {
	return res.LastInsertId()
}

func (NullInterceptor) ResultRowsAffected(res driver.Result) (int64, error) {
	return res.RowsAffected()
}

func (NullInterceptor) RowsNext(rows driver.Rows, dest []driver.Value) error {
	return rows.Next(dest)
}

func (NullInterceptor) RowsClose(rows driver.Rows) error {
	return rows.Close()
}

func (NullInterceptor) StmtExecContext(ctx context.Context, stmt *Stmt, args []driver.NamedValue) (driver.Result, error) {
	return stmt.ExecContext(ctx, args)
}

func (NullInterceptor) StmtQueryContext(ctx context.Context, stmt *Stmt, args []driver.NamedValue) (driver.Rows, error) {
	return stmt.QueryContext(ctx, args)
}

func (NullInterceptor) StmtClose(stmt *Stmt) error {
	return stmt.Close()
}

func (NullInterceptor) TxCommit(tx driver.Tx) error {
	return tx.Commit()
}

func (NullInterceptor) TxRollback(tx driver.Tx) error {
	return tx.Rollback()
}
