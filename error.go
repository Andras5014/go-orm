package go_orm

import "github.com/Andras5014/go-orm/internal/errs"

// 内部错误暴露在外面
var (
	ErrNoRows = errs.ErrNoRows
)
