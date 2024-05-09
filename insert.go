package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"reflect"
	"strings"
)

type Inserter[T any] struct {
	values []*T
	db     *DB
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
	}
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
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
	for idx, field := range m.Fields {
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

	args := make([]any, 0, len(i.values)*len(m.Fields))
	for index, val := range i.values {
		if index > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(")
		for idx, field := range m.Fields {
			if idx > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			args = append(args, arg)
		}
		sb.WriteString(")")
	}

	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: args,
	}, nil
}
