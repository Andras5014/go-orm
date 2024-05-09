package go_orm

import "strings"

type Inserter[T any] struct {
	values []*T
	db     *DB
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{}
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}
func (i *Inserter[T]) Build() (*Query, error) {
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
	cnt := 0
	for _, field := range m.FieldMap {
		if cnt > 0 {
			sb.WriteString("`,")
		}
		sb.WriteString("`")
		sb.WriteString(field.ColName)
		sb.WriteString("`")
		cnt++
	}
	sb.WriteString(")")
	return &Query{
		SQL:  sb.String(),
		Args: nil,
	}, nil
}
