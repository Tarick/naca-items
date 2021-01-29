// Package tracing reimplements opentracing hook from
// https://github.com/99designs/gqlgen-contrib/blob/master/gqlopentracing/tracer.go
// using non-global tracer, but one from dependency injection
package tracing

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otLog "github.com/opentracing/opentracing-go/log"
)

// New returns opentracing tracer wrapper with necessary methods
func New(tracer opentracing.Tracer) *gqlTracer {
	return &gqlTracer{tracer: tracer}
}

type gqlTracer struct {
	tracer opentracing.Tracer
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
} = &gqlTracer{}

func (a gqlTracer) ExtensionName() string {
	return "OpenTracing"
}

func (a gqlTracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (a *gqlTracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	oc := graphql.GetOperationContext(ctx)
	var spanName string = oc.OperationName
	if spanName == "" {
		// Raw query is usually big and looks ugly in tracing dashboard interface, but we need to use it somehow if Operation Name is not specified in query
		spanName = oc.RawQuery
	}
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, a.tracer, spanName)
	defer span.Finish()

	ext.SpanKind.Set(span, "server")
	ext.Component.Set(span, "gqlgen")
	span.SetTag("gqlQuery", oc.RawQuery)

	return next(ctx)
}

func (a *gqlTracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, a.tracer, fc.Object+"_"+fc.Field.Name)
	defer span.Finish()
	span.SetTag("resolver.object", fc.Object)
	span.SetTag("resolver.field", fc.Field.Name)

	res, err := next(ctx)

	errList := graphql.GetFieldErrors(ctx, fc)
	if len(errList) != 0 {
		ext.Error.Set(span, true)
		span.LogFields(
			otLog.String("event", "error"),
		)

		for idx, err := range errList {
			span.LogFields(
				otLog.String(fmt.Sprintf("error.%d.message", idx), err.Error()),
				otLog.String(fmt.Sprintf("error.%d.kind", idx), fmt.Sprintf("%T", err)),
			)
		}
	}

	return res, err
}
