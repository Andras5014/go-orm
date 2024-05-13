package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
)

var (
	DialectMySQL      = &mysqlDialect{}
	DialectSQLite     = &sqliteDialect{}
	DialectPostgreSQL = &postgresDialect{}
)

type Dialect interface {
	// quoter 解决引号问题
	quoter() byte

	buildUpsert(b *builder, upsert *Upsert) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildUpsert(b *builder, upsert *Upsert) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (m mysqlDialect) quoter() byte {
	return '`'
}
func (m mysqlDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range upsert.assigns {
		if idx > 0 {
			b.sb.WriteString(",")
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.addArg(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteString(")")

		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

type sqliteDialect struct {
	standardSQL
}

func (s sqliteDialect) quoter() byte {
	return '`'
}
func (s sqliteDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON CONFLICT (")
	for i, col := range upsert.conflictColumns {
		if i > 0 {
			b.sb.WriteString(",")
		}
		err := b.buildColumn(Column{name: col})
		if err != nil {
			return err
		}
	}
	b.sb.WriteString(") DO UPDATE SET ")
	for i, assign := range upsert.assigns {
		if i > 0 {
			b.sb.WriteString(",")
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.addArg(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=EXCLUDED.")
			b.quote(fd.ColName)
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

type postgresDialect struct {
	standardSQL
}
