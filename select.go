package go_orm

import (
	"context"
	"github.com/Andras5014/go-orm/internal/errs"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
	model *Model
	where []Predicate
	sb    *strings.Builder
	args  []any
	db    *DB
	//r *registry
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		sb: &strings.Builder{},
		db: db,
	}
}
func (s *Selector[T]) Build() (*Query, error) {

	var err error

	s.model, err = s.db.r.Register(new(T))

	if err != nil {
		return nil, err
	}
	sb := s.sb
	sb.WriteString("SELECT * FROM ")

	if s.table == "" {
		sb.WriteByte('`')
		sb.WriteString(s.model.tableName)
		sb.WriteByte('`')
	} else {

		sb.WriteString(s.table)

	}

	if len(s.where) > 0 {
		sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}

		if err := s.buildExpression(p); err != nil {
			return nil, err
		}

	}
	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: s.args,
	}, nil

}

func (s *Selector[T]) buildExpression(expr Expression) error {

	switch exp := expr.(type) {
	case nil:
		return nil
	case Predicate:
		// 构建 p.left
		// 构建 p.op
		// 构建 p.right

		_, ok := exp.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')
		_, ok = exp.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

	case Column:

		fd, ok := s.model.fieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')

	case value:
		s.addArg(exp.arg)
		s.sb.WriteString("?")
	default:
		return errs.NewErrUnsupportedExpr(exp)
	}
	return nil
}

func (s *Selector[T]) addArg(val any) *Selector[T] {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, val)
	return s
}
func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}
func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	// 构造sql失败
	if err != nil {
		return nil, err
	}
	// 发起查询, 处理结果集
	db := s.db.db
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	// 查询错误
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, ErrNoRows
	}

	// 拿到 select 出来的列
	cs, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	tp := new(T)

	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		val := reflect.New(fd.typ)
		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())

	}
	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}
	tpValueElem := reflect.ValueOf(tp).Elem()
	for i, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		tpValueElem.FieldByName(fd.goName).Set(valElems[i])

	}
	return tp, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) (*[]T, error) {
	//q, err := s.Build()
	//if err != nil {
	//	return nil, err
	//}
	//// 执行查询, 处理结果集
	//db := s.db.db
	//rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	//for rows.Next() {
	//
	//}
	panic("implement me")
}
