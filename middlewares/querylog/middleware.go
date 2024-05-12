package querylog

import (
	"context"
	go_orm "github.com/Andras5014/go-orm"
	"log"
)

type MiddlewareBuilder struct {
	logFunc func(query string, args []any)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(query string, args []any) {
			log.Printf("sql: %s, args: %v", query, args)
		},
	}
}

func (m *MiddlewareBuilder) LogFunc(logFunc func(query string, args []any)) *MiddlewareBuilder {
	m.logFunc = logFunc
	return m
}
func (m *MiddlewareBuilder) Build() go_orm.Middleware {
	return func(next go_orm.Handler) go_orm.Handler {
		return func(ctx context.Context, qc *go_orm.QueryContext) *go_orm.QueryResult {
			q, err := qc.Builder.Build()
			if err != nil {
				return &go_orm.QueryResult{
					Err: err,
				}
			}
			m.logFunc(q.SQL, q.Args)
			res := next(ctx, qc)
			return res
		}
	}
}
