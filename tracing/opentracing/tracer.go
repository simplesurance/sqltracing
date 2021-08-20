// Package opentracing provides an opentracing-go Tracer that is compatible
// with the sqltracing.Tracer interface.
package opentracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/simplesurance/sqltracing"
)

// DefaultTracingTags are the tags that are added by default to all traces.
var DefaultTracingTags = opentracing.Tags{
	string(ext.Component): "sqltracing",
	string(ext.DBType):    "sql",
	string(ext.SpanKind):  string(ext.SpanKindRPCClientEnum),
}

type tracer struct {
	defaultTags  opentracing.Tags
	traceOrphans bool
	getTracerFn  func() opentracing.Tracer
}

type span struct {
	*tracer
	span opentracing.Span
}

// Opt is a type for options that can be passed to NewTracer.
type Opt func(*tracer)

// WithTracingTags is an option for NewTracer() to set the tags that are
// applied to all traces created by StartSpan().
func WithTracingTags(tags opentracing.Tags) Opt {
	return func(t *tracer) {
		t.defaultTags = tags
	}
}

// WithoutTracingOrphans is an option for NewTracer() to disable tracing spans
// if no parent span exist in the context.
func WithoutTracingOrphans() Opt {
	return func(t *tracer) {
		t.traceOrphans = false
	}
}

// WithTracer is an option for NewTracer() to use a custom function to
// retrieve the opentracing Tracer to use.
func WithTracer(fn func() opentracing.Tracer) Opt {
	return func(t *tracer) {
		t.getTracerFn = fn
	}
}

// NewTracer returns a tracer that will create spans via opentracing-go.
// When no options are specified, opentracing.GlobalTracer is used as default
// Tracer, DefaultTracingTags are used as tags and TraceOrphans is enabled.
func NewTracer(opts ...Opt) sqltracing.Tracer {
	tr := tracer{
		traceOrphans: true,
		defaultTags:  DefaultTracingTags,
		getTracerFn:  opentracing.GlobalTracer,
	}

	for _, opt := range opts {
		opt(&tr)
	}

	return &tr
}

func (t *tracer) StartSpan(ctx context.Context, name string) (sqltracing.Span, context.Context) {
	parent := opentracing.SpanFromContext(ctx)
	if parent != nil {
		otSpan := parent.Tracer().StartSpan(
			name,
			opentracing.ChildOf(parent.Context()),
			t.defaultTags,
		)
		return &span{span: otSpan, tracer: t}, opentracing.ContextWithSpan(ctx, otSpan)
	}

	if !t.traceOrphans {
		return &span{span: nil, tracer: t}, ctx
	}

	otSpan := t.getTracerFn().StartSpan(name, t.defaultTags)
	return &span{span: otSpan, tracer: t}, opentracing.ContextWithSpan(ctx, otSpan)
}

func (s *span) SetTag(k, v string) {
	if s.span == nil {
		return
	}

	s.span.SetTag(k, v)
}

func (s *span) SetTags(kvs map[string]string) {
	if s.span == nil {
		return
	}

	for k, v := range kvs {
		s.span.SetTag(k, v)
	}
}

func (s *span) SetError(err error) {
	if s.span == nil {
		return
	}

	ext.LogError(s.span, err)
}

func (s *span) Finish() {
	if s.span == nil {
		return
	}

	s.span.Finish()
}
