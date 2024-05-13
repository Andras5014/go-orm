package go_orm

// Expression 标记接口，代表表达式
type Expression interface {
	expr()
}

// RawExpr 表示原生表达式
type RawExpr struct {
	raw  string
	args []any
}

func (r RawExpr) assign() {
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

func (r RawExpr) selectable() {
}
func (r RawExpr) expr() {
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
