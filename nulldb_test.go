package sqltracing_test

import (
	"context"
	"database/sql/driver"
	"io"
)

type nullDriver struct {
	con driver.Conn
}

func (d *nullDriver) Open(_ string) (driver.Conn, error) {
	return d.con, nil
}

type nullStmt struct{}

func (s *nullStmt) Close() error {
	return nil
}

func (s *nullStmt) NumInput() int {
	return 0
}

func (s *nullStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return &nullRows{}, nil
}

func (s *nullStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return nil, nil
}

type nullRows struct{}

type nullTx struct{}

func (r *nullRows) Close() error {
	return nil
}

func (r *nullRows) Columns() []string {
	return []string{"1"}
}

func (r *nullRows) Next(_ []driver.Value) error {
	return io.EOF
}

func (t *nullTx) Commit() error {
	return nil
}

func (t *nullTx) Rollback() error {
	return nil
}

type nullCon struct{}

func (c *nullCon) Prepare(_ string) (driver.Stmt, error) {
	return &nullStmt{}, nil
}

func (c *nullCon) Close() error { return nil }

func (c *nullCon) Begin() (driver.Tx, error) {
	return &nullTx{}, nil
}

func (c *nullCon) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	return &nullRows{}, nil
}

func (c *nullCon) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (c *nullCon) Ping(_ context.Context) error {
	return nil
}
