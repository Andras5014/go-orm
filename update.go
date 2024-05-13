package go_orm

import (
	"context"
	"github.com/Andras5014/go-orm/internal/errs"
)

type Updater[T any] struct {
	builder
	table   string
	assigns []Assignable
	val     *T
	where   []Predicate
	sess    Session
}

func NewUpdater[T any](sess Session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		sess: sess,
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
	}
}

func (u *Updater[T]) Build() (*Query, error) {
	m, err := u.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	u.model = m

	u.sb.WriteString("UPDATE ")
	if u.table == "" {
		u.quote(m.TableName)
	} else {
		u.sb.WriteString(u.table)
	}

	u.sb.WriteString(" SET ")
	if len(u.assigns) == 0 {
		return nil, errs.ErrNoUpdatedColumns
	}

	for i, s := range u.assigns {
		if i > 0 {
			u.sb.WriteByte(',')
		}
		switch v := s.(type) {
		case Assignment:
			fd, ok := m.FieldMap[v.col]
			if !ok {
				return nil, errs.NewErrUnknownField(v.col)
			}
			u.quote(fd.ColName)
			u.sb.WriteString(" = ?")
			u.addArg(v.val)
		case RawExpr:
			u.sb.WriteString(v.raw)
			u.addArg(v.args...)
		default:
			return nil, errs.NewErrUnsupportedAssignableType(s)
		}
	}

	if len(u.where) > 0 {
		u.sb.WriteString(" WHERE ")
		if err = u.buildPredicates(u.where); err != nil {
			return nil, err
		}
	}

	u.sb.WriteByte(';')
	return &Query{
		SQL:  u.sb.String(),
		Args: u.args,
	}, nil

}

func (u *Updater[T]) Set(assignments ...Assignable) *Updater[T] {
	u.assigns = append(u.assigns, assignments...)
	return u
}

func (u *Updater[T]) Where(predicates Predicate) *Updater[T] {
	u.where = append(u.where, predicates)
	return u
}

func (u *Updater[T]) Exec(ctx context.Context) Result {
	q, err := u.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := u.sess.execContext(ctx, q.SQL, q.Args...)
	return Result{err: err, res: res}
}
