package sqlmw

import (
	"context"
	"database/sql/driver"
)

type wrappedStmt struct {
	intr   Interceptor
	parent driver.Stmt
	conn   wrappedConn
}

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Stmt              = wrappedStmt{}
	_ driver.StmtExecContext   = wrappedStmt{}
	_ driver.StmtQueryContext  = wrappedStmt{}
	_ driver.ColumnConverter   = wrappedStmt{}
	_ driver.NamedValueChecker = wrappedStmt{}
)

func (s wrappedStmt) Close() (err error) {
	return s.intr.StmtClose(&Stmt{Stmt: s.parent})
}

func (s wrappedStmt) NumInput() int {
	return s.parent.NumInput()
}

func (s wrappedStmt) Exec(args []driver.Value) (res driver.Result, err error) {
	res, err = s.parent.Exec(args)
	if err != nil {
		return nil, err
	}
	return wrappedResult{intr: s.intr, parent: res}, nil
}

func (s wrappedStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	rows, err = s.parent.Query(args)
	if err != nil {
		return nil, err
	}
	return wrappedRows{intr: s.intr, parent: rows}, nil
}

func (s wrappedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	wrappedParent := Stmt{Stmt: s.parent}
	res, err = s.intr.StmtExecContext(ctx, &wrappedParent, args)
	if err != nil {
		return nil, err
	}
	return wrappedResult{intr: s.intr, parent: res}, nil
}

func (s wrappedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	wrappedParent := Stmt{Stmt: s.parent}
	rows, err = s.intr.StmtQueryContext(ctx, &wrappedParent, args)
	if err != nil {
		return nil, err
	}
	return wrappedRows{intr: s.intr, parent: rows}, nil
}

func (s wrappedStmt) ColumnConverter(idx int) driver.ValueConverter {
	if converter, ok := s.parent.(driver.ColumnConverter); ok {
		return converter.ColumnConverter(idx)
	}

	return driver.DefaultParameterConverter
}

// Stmt makes a Stmt compatible with the StmtExecContext and
// StmtQueryContext interfaces.
// If the wrapped Stmt does not support those methods, StmtExec and StmtQuery
// are called as fallback.
type Stmt struct {
	driver.Stmt
}

// Parent returns the original Stmt that was created and returned by
// ConnPrepareContext.
func (s *Stmt) Parent() driver.Stmt {
	return s.Stmt
}

func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	if stmtQueryContext, ok := s.Stmt.(driver.StmtQueryContext); ok {
		return stmtQueryContext.QueryContext(ctx, args)
	}
	// Fallback implementation
	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return s.Stmt.Query(dargs)
}

func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	if stmtExecContext, ok := s.Stmt.(driver.StmtExecContext); ok {
		return stmtExecContext.ExecContext(ctx, args)
	}
	// Fallback implementation
	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return s.Stmt.Exec(dargs)
}

func (s wrappedStmt) CheckNamedValue(v *driver.NamedValue) error {
	if checker, ok := s.parent.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(v)
	}

	if checker, ok := s.conn.parent.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(v)
	}

	return defaultCheckNamedValue(v)
}
