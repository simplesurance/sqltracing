package sqltracing

import (
	"context"
	"database/sql/driver"
)

type tracedTx struct {
	driver.Tx
	ctx                context.Context
	parentSpanFinishFn func(err error)
}

func newTracedTx(ctx context.Context, parentSpanFinishFn func(error), tx driver.Tx) *tracedTx {
	return &tracedTx{
		Tx:                 tx,
		ctx:                ctx,
		parentSpanFinishFn: parentSpanFinishFn,
	}
}
