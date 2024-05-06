package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly = errors.New("orm: only support one pointer")
	ErrNoRows      = errors.New("orm: no rows in result set")
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

func NewErrUnknownColumn(c string) error {
	return fmt.Errorf("orm: unknown column: %s", c)
}
