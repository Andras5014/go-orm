package opentelemetry

import (
	"context"
	"fmt"
	go_orm "github.com/Andras5014/go-orm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/Andras5014/go-orm/middlewares/opentelemetry"
)

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m *MiddlewareBuilder) Builder() go_orm.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next go_orm.Handler) go_orm.Handler {
		return func(ctx context.Context, qc *go_orm.QueryContext) *go_orm.QueryResult {

			// span name: select-test_model
			tbl := qc.Model.TableName
			spanCtx, span := m.Tracer.Start(ctx, fmt.Sprintf("%s-%s", qc.Type, tbl))
			defer span.End()
			q, _ := qc.Builder.Build()
			if q != nil {
				span.SetAttributes(attribute.String("sql", q.SQL))
			}
			span.SetAttributes(attribute.String("table", tbl))
			res := next(spanCtx, qc)
			if res.Err != nil {
				span.RecordError(res.Err)
			}
			return res
		}
	}
}
