package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/Andras5014/go-orm/model"
	"reflect"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}
type OnDuplicateKey struct {
	assigns []Assignable
}

func (o OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.OnDuplicateKey = &OnDuplicateKey{
		assigns: assigns,
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
	db             *DB
	OnDuplicateKey *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
	}
}

func (i *Inserter[T]) onDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
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
	m, err := i.db.r.Get(i.values[0])
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
	for index, val := range i.values {
		if index > 0 {
			i.sb.WriteString(",")
		}
		i.sb.WriteString("(")
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteString(",")
			}
			i.sb.WriteString("?")
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			i.addArg(arg)
		}
		i.sb.WriteString(")")
	}
	if i.OnDuplicateKey != nil {
		err := i.dialect.buildOnDuplicateKey(&i.builder, i.OnDuplicateKey)
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
