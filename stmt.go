package sqltracing

import (
	"context"
	"database/sql/driver"
)

type tracedStmt struct {
	driver.Stmt
	ctx                context.Context
	parentSpanFinishFn func(err error)
}

func newTracedStmt(ctx context.Context, parentSpanFinishFn func(error), stmt driver.Stmt) *tracedStmt {
	return &tracedStmt{
		Stmt:               stmt,
		ctx:                ctx,
		parentSpanFinishFn: parentSpanFinishFn,
	}
}
