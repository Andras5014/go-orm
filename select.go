package go_orm

import (
	"context"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/Andras5014/go-orm/model"
	"strings"
)

// Selectable 是一个标记接口
// 查找的列，聚合函数
type Selectable interface {
	selectable()
}
type Selector[T any] struct {
	table   string
	model   *model.Model
	where   []Predicate
	columns []Selectable
	sb      *strings.Builder
	args    []any
	db      *DB
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
	sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	sb.WriteString(" FROM ")

	if s.table == "" {
		sb.WriteByte('`')
		sb.WriteString(s.model.TableName)
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
		return s.buildColumn(exp.name)
	case value:
		s.addArg(exp.arg)
		s.sb.WriteString("?")
	default:
		return errs.NewErrUnsupportedExpr(exp)
	}
	return nil
}
func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
	}

	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c.name)
			if err != nil {
				return err
			}
		case Aggregate:
			s.sb.WriteString(c.fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(c.arg)
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
		}

	}

	return nil
}

func (s *Selector[T]) buildColumn(c string) error {
	fd, ok := s.model.FieldMap[c]
	if !ok {
		return errs.NewErrUnknownField(c)
	}
	s.sb.WriteByte('`')
	s.sb.WriteString(fd.ColName)
	s.sb.WriteByte('`')
	return nil
}
func (s *Selector[T]) addArg(val any) *Selector[T] {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, val)
	return s
}

//	func (s *Selector[T]) Select(columns ...string) *Selector[T] {
//		s.columns = columns
//		return s
//	}
func (s *Selector[T]) Select(columns ...Selectable) *Selector[T] {
	s.columns = columns
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

	tp := new(T)
	val := s.db.creator(s.model, tp)
	err = val.SetColumns(rows)
	return tp, err

}

//func (s *Selector[T]) GetV1(ctx context.Context) (*T, error) {
//	q, err := s.Build()
//	// 构造sql失败
//	if err != nil {
//		return nil, err
//	}
//	// 发起查询, 处理结果集
//	db := s.db.db
//	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
//	// 查询错误
//	if err != nil {
//		return nil, err
//	}
//	if !rows.Next() {
//		return nil, ErrNoRows
//	}
//
//}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	// 执行查询, 处理结果集
	db := s.db.db
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	var res []*T
	for rows.Next() {
		tp := new(T)
		val := s.db.creator(s.model, tp)
		err = val.SetColumns(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, tp)
	}
	return res, nil
}
