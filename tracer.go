package sqltracing

import "context"

// Tracer defines the required methods of a Tracer implementation
type Tracer interface {
	// StartSpan extracts the parent span from ctx and starts a child span
	// called spanName. It returns the child span and a new context that
	// contains the child span.
	StartSpan(ctx context.Context, spanName string) (Span, context.Context)
}

// Span is part of the interface needed to be implemented by any tracing implementation we use
type Span interface {
	// SetTag sets the tags with identifier k to value v.
	SetTag(k, v string)
	// SetTags sets the tags of the span to given key-value map.
	SetTags(kvs map[string]string)
	// SetError adds information of the error to the span.
	SetError(err error)
	// Finish finishes the span.
	Finish()
}
