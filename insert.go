package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/Andras5014/go-orm/model"
	"reflect"
	"strings"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}
type OnDuplicateKey[T any] struct {
	assigns []Assignable
}

func (o OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.OnDuplicateKey = &OnDuplicateKey[T]{
		assigns: assigns,
	}
	return o.i
}

type Assignable interface {
	assign()
}
type Inserter[T any] struct {
	values         []*T
	columns        []string
	db             *DB
	OnDuplicateKey *OnDuplicateKey[T]
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
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
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	sb.WriteByte('`')
	sb.WriteString(m.TableName)
	sb.WriteByte('`')
	// 指定列的顺序
	sb.WriteString(" (")

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
			sb.WriteString(",")
		}
		sb.WriteString("`")
		sb.WriteString(field.ColName)
		sb.WriteString("`")
	}
	sb.WriteString(")")

	// 拼接 VALUES
	sb.WriteString(" VALUES ")

	args := make([]any, 0, len(i.values)*len(fields))
	for index, val := range i.values {
		if index > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(")
		for idx, field := range fields {
			if idx > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			args = append(args, arg)
		}
		sb.WriteString(")")
	}
	if i.OnDuplicateKey != nil {
		sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		for idx, assign := range i.OnDuplicateKey.assigns {
			if idx > 0 {
				sb.WriteString(",")
			}
			switch a := assign.(type) {
			case Assignment:
				fd, ok := m.FieldMap[a.col]
				if !ok {
					return nil, errs.NewErrUnknownField(a.col)
				}
				sb.WriteString("`")
				sb.WriteString(fd.ColName)
				sb.WriteString("`")
				sb.WriteString("=?")
				args = append(args, a.val)
			case Column:
				fd, ok := m.FieldMap[a.name]
				if !ok {
					return nil, errs.NewErrUnknownField(a.name)
				}
				sb.WriteString("`")
				sb.WriteString(fd.ColName)
				sb.WriteString("`")
				sb.WriteString("=VALUES(")
				sb.WriteString("`")
				sb.WriteString(fd.ColName)
				sb.WriteString("`")
				sb.WriteString(")")

			default:
				return nil, errs.NewErrUnsupportedAssignable(assign)
			}
		}
	}
	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: args,
	}, nil
}
