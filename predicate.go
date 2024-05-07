package go_orm

type op string

const (
	opEq  op = "="
	opLt  op = "<"
	opGt  op = ">"
	opNot op = "NOT"
	opAnd op = "AND"
	opOr  op = "OR"
)

func (o op) String() string {
	return string(o)
}

// Expression 标记接口，代表表达式
type Expression interface {
	expr()
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (Predicate) expr() {}

type value struct {
	arg any
}

func (value) expr() {}

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLt,
		right: value{arg: arg},
	}
}
func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGt,
		right: value{arg: arg},
	}
}

func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}
