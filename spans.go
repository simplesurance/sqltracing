package sqltracing

import (
	"context"
	"errors"
)

// DBStatementTagKey is the name of the tracing that contains db query
// statements.
const DBStatementTagKey = "db.statement"

func (d *Interceptor) startSpan(ctx context.Context, opName SQLOp, query string, whitelistedErr ...error) (func(err error), context.Context) {
	if d.opIsExcluded(opName) {
		return func(_ error) {}, ctx
	}

	span, ctx := d.tracer.StartSpan(ctx, opName.String())

	if query != "" {
		span.SetTag(DBStatementTagKey, query)
	}

	return spanFinishFunc(span, whitelistedErr...), ctx
}

func spanFinishFunc(span Span, whitelistedErr ...error) func(err error) {
	return func(err error) {
		if err != nil && !errisOneOf(err, whitelistedErr) {
			span.SetError(err)
		}

		span.Finish()
	}
}

func errisOneOf(err error, targets []error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}
