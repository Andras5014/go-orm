package go_orm

import (
	"context"
	"github.com/Andras5014/go-orm/internal/errs"
)

// Selectable 是一个标记接口
// 查找的列，聚合函数
type Selectable interface {
	selectable()
}
type Selector[T any] struct {
	builder
	table    string
	where    []Predicate
	having   []Predicate
	columns  []Selectable
	groupBys []Column
	orderBys []OrderBy
	offset   int
	limit    int

	sess Session
	//r *registry
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}
func (s *Selector[T]) Build() (*Query, error) {
	if s.model == nil {
		var err error
		s.model, err = s.r.Get(new(T))
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteString("SELECT ")
	if err := s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")

	if s.table == "" {
		s.quote(s.model.TableName)
	} else {
		s.sb.WriteString(s.table)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		if err := s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}
	if len(s.groupBys) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, column := range s.groupBys {
			if i > 0 {
				s.sb.WriteString(", ")
			}
			if err := s.buildColumn(column); err != nil {
				return nil, err
			}
		}
	}
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		if err := s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}
	// limit x,y x偏移量y目标量
	// 从偏移量x往后取y条数据
	// limit y offset x
	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArg(s.limit)
	}
	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArg(s.offset)
	}
	if len(s.orderBys) > 0 {
		s.sb.WriteString(" ORDER BY ")
		if err := s.buildOrderBy(); err != nil {
			return nil, err
		}
	}
	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil

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
			err := s.buildColumn(c)
			if err != nil {
				return err
			}
		case Aggregate:
			err := s.buildAggregate(c, true)
			if err != nil {
				return err
			}
		case RawExpr:
			s.sb.WriteString(c.raw)
			s.addArg(c.args...)
		}

	}

	return nil
}

func (s *Selector[T]) buildAggregate(a Aggregate, useAlias bool) error {
	s.sb.WriteString(a.fn)
	s.sb.WriteByte('(')
	err := s.buildColumn(Column{name: a.arg})
	if err != nil {
		return err
	}
	s.sb.WriteByte(')')
	if a.alias != "" {
		s.sb.WriteString(" AS `")
		s.sb.WriteString(a.alias)
		s.sb.WriteByte('`')
	}
	return nil
}
func (s *Selector[T]) buildColumn(c Column) error {
	fd, ok := s.model.FieldMap[c.name]
	if !ok {
		return errs.NewErrUnknownField(c.name)
	}
	s.quote(fd.ColName)
	if c.alias != "" {
		s.sb.WriteString(" AS `")
		s.sb.WriteString(c.alias)
		s.sb.WriteByte('`')
	}
	return nil
}
func (s *Selector[T]) buildOrderBy() error {

	for i, ob := range s.orderBys {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		err := s.buildColumn(ob.col)
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(ob.order)
	}
	return nil
}

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
func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}
func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBys = cols
	return s
}
func (s *Selector[T]) OrderBy(orderBys ...OrderBy) *Selector[T] {
	s.orderBys = orderBys
	return s
}
func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}
func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {

	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Model:   s.model,
		Type:    "SELECT",
		Builder: s,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	// 执行查询, 处理结果集
	rows, err := s.sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	var res []*T
	for rows.Next() {
		tp := new(T)
		val := s.creator(s.model, tp)
		err = val.SetColumns(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, tp)
	}
	return res, nil
}

type OrderBy struct {
	col   Column
	order string
}

func Asc(col string) OrderBy {
	return OrderBy{
		col:   Column{name: col},
		order: "ASC",
	}
}
func Desc(col string) OrderBy {
	return OrderBy{
		col:   Column{name: col},
		order: "DESC",
	}
}
