package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"strings"
)

type builder struct {
	core
	sb     strings.Builder
	args   []any
	quoter byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

// buildPredicate 构造谓词（条件）
func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

// buildExpression 递归构造条件表达式
func (b *builder) buildExpression(expr Expression) error {

	switch exp := expr.(type) {
	case nil:
		return nil
	case Predicate:
		// 构建 p.left
		// 构建 p.op
		// 构建 p.right

		_, ok := exp.left.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
		if exp.op != "" {
			b.sb.WriteByte(' ')
			b.sb.WriteString(exp.op.String())
			b.sb.WriteByte(' ')
		}

		_, ok = exp.right.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}

	case Column:
		exp.alias = ""
		return b.buildColumn(Column{name: exp.name})
	case value:
		b.addArg(exp.arg)
		b.sb.WriteString("?")
	case RawExpr:
		b.sb.WriteString(exp.raw)
		b.addArg(exp.args...)
	case Aggregate:
		b.sb.WriteString(exp.fn)
		b.sb.WriteString("(`")
		fd, ok := b.model.FieldMap[exp.arg]
		if !ok {
			return errs.NewErrUnknownColumn(exp.arg)
		}
		b.sb.WriteString(fd.ColName)
		b.sb.WriteString("`)")

	default:
		return errs.NewErrUnsupportedExpr(exp)
	}
	return nil
}
func (b *builder) buildColumn(c Column) error {
	fd, ok := b.model.FieldMap[c.name]
	if !ok {
		return errs.NewErrUnknownColumn(c.name)
	}
	b.quote(fd.ColName)
	return nil
}

func (b *builder) addArg(args ...any) {
	if len(args) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}
