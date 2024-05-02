package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly = errors.New("orm: only support one pointer")
)

func NewErrUnsupportedExpr(expr any) error {
	return fmt.Errorf("orm: unsupported expression: %s", expr)
}
func NewErrUnknownField(name string) error {
	return fmt.Errorf("orm: unknown field: %s", name)
}

func NewErrInvalidTagContent(pair string) error {
	return fmt.Errorf("orm: invalid tag content: %s", pair)
}
