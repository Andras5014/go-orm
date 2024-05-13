package go_orm

import "context"

type Deleter[T any] struct {
	builder
	table string
	where []Predicate
	sess  Session
}

// NewDeleter 开始构建一个 DELETE 查询
func NewDeleter[T any](sess Session) *Deleter[T] {
	c := sess.getCore()
	return &Deleter[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}
func (d *Deleter[T]) Build() (*Query, error) {
	m, err := d.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	d.model = m

	d.sb.WriteString("DELETE FROM ")
	// 表名 如果没有指定表名，则使用类型名
	if d.table == "" {
		d.quote(d.model.TableName)
	} else {
		// 自己指定表名，不会自动加反引号， 因为可能是 db.table 这种形式
		d.sb.WriteString(d.table)
	}

	// 条件构造
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")

		if err := d.buildPredicates(d.where); err != nil {
			return nil, err
		}

	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil

}

func (d *Deleter[T]) Form(table string) *Deleter[T] {
	d.table = table
	return d
}
func (d *Deleter[T]) Where(prs ...Predicate) *Deleter[T] {
	d.where = prs
	return d
}

// Exec sql
func (d *Deleter[T]) Exec(ctx context.Context) Result {
	query, err := d.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := d.sess.execContext(ctx, query.SQL, query.Args)
	return Result{
		err: err,
		res: res,
	}
}
