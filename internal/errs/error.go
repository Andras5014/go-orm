package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly      = errors.New("orm: only support one pointer")
	ErrNoRows           = errors.New("orm: no rows in result set")
	ErrInsertZeroRow    = errors.New("orm: insert zero row")
	ErrNoUpdatedColumns = errors.New("orm: no updated columns")
)

// NewErrFailedToRollback bizErr 是业务错误，rbErr 是回滚错误，panicked 是是否在回滚时发生 panic
func NewErrFailedToRollback(bizErr error, rbErr error, panicked bool) error {
	return fmt.Errorf("orm: failed to rollback: %w, %w, %t", bizErr, rbErr, panicked)
}
func NewErrUnsupportedExpr(expr any) error {
	return fmt.Errorf("orm: unsupported expression: %s", expr)
}
func NewErrUnknownField(name string) error {
	return fmt.Errorf("orm: unknown field: %s", name)
}

func NewErrInvalidTagContent(pair string) error {
	return fmt.Errorf("orm: invalid tag content: %s", pair)
}

func NewErrUnknownColumn(c string) error {
	return fmt.Errorf("orm: unknown column: %s", c)
}
func NewErrUnsupportedAssignable(assign any) error {
	return fmt.Errorf("orm: unsupported assignable: %s", assign)
}
func NewErrUnsupportedAssignableType(assign any) error {
	return fmt.Errorf("orm: unsupported assignable type: %s", assign)
}
