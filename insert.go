package go_orm

import (
	"context"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/Andras5014/go-orm/model"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}
type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

// ConflictColumns 中间方法
func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.OnDuplicateKey = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Assignable interface {
	assign()
}
type Inserter[T any] struct {
	builder
	values         []*T
	columns        []string
	OnDuplicateKey *Upsert
	sess           Session
}

func NewInserter[T any](sess Session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (i *Inserter[T]) onDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

// Values 指定插入的数据
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}
func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}

	i.sb.WriteString("INSERT INTO ")
	m, err := i.r.Get(i.values[0])
	i.model = m
	if err != nil {
		return nil, err
	}
	i.quote(m.TableName)
	// 指定列的顺序
	i.sb.WriteString(" (")

	fields := m.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := m.FieldMap[fd]
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}
	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteString(",")
		}
		i.quote(field.ColName)
	}
	i.sb.WriteString(")")

	// 拼接 VALUES
	i.sb.WriteString(" VALUES ")

	i.args = make([]any, 0, len(i.values)*len(fields))
	for index, v := range i.values {
		if index > 0 {
			i.sb.WriteString(",")
		}
		i.sb.WriteString("(")
		val := i.creator(i.model, v)
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteString(",")
			}
			i.sb.WriteString("?")
			arg, err := val.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArg(arg)
		}
		i.sb.WriteString(")")
	}
	if i.OnDuplicateKey != nil {
		err := i.dialect.buildUpsert(&i.builder, i.OnDuplicateKey)
		if err != nil {
			return nil, err
		}
	}
	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

func (i Inserter[T]) Exec(ctx context.Context) Result {
	q, err := i.Build()
	if err != nil {
		return Result{
			err: err,
		}
	}
	res, err := i.sess.execContext(ctx, q.SQL, q.Args...)
	return Result{
		res: res,
		err: err,
	}
}
