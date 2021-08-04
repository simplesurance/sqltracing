package sqltracing

import (
	"context"
	"database/sql/driver"
)

type tracedRows struct {
	driver.Rows
	ctx                context.Context
	parentSpanFinishFn func(err error)
}

func newTracedRows(ctx context.Context, parentSpanFinishFn func(error), rows driver.Rows) *tracedRows {
	return &tracedRows{
		Rows:               rows,
		ctx:                ctx,
		parentSpanFinishFn: parentSpanFinishFn,
	}
}
