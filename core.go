package go_orm

import (
	"context"
	"github.com/Andras5014/go-orm/internal/valuer"
	"github.com/Andras5014/go-orm/model"
)

type core struct {
	model       *model.Model
	dialect     Dialect
	creator     valuer.Creator
	r           model.Registry
	middlewares []Middleware
}

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		root = c.middlewares[i](root)
	}
	return root(ctx, qc)

}
func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	// 构造sql失败
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// 发起查询, 处理结果集
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	// 查询错误
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	if !rows.Next() {
		return &QueryResult{
			Err: ErrNoRows,
		}
	}

	tp := new(T)
	val := c.creator(c.model, tp)
	err = val.SetColumns(rows)
	return &QueryResult{
		Err:    err,
		Result: tp,
	}
}
func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, c, qc)
	}
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		root = c.middlewares[i](root)
	}
	return root(ctx, qc)

}
func execHandler(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
			Result: Result{
				err: err,
			},
		}
	}
	res, err := sess.execContext(ctx, q.SQL, q.Args...)
	return &QueryResult{
		Err: err,
		Result: Result{
			err: err,
			res: res,
		},
	}
}
