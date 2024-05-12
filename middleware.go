package go_orm

import (
	"context"
	"github.com/Andras5014/go-orm/model"
)

type QueryContext struct {
	// 查询类型, select, insert, delete, update
	Type string
	// 查询语句
	Builder QueryBuilder

	Model *model.Model
}

type QueryResult struct {
	// Result 查询结果在不同查询类型下不同
	// select: *T or []*T
	Result any
	// Err 查询错误
	Err error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type Middleware func(next Handler) Handler
