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

	buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (m mysqlDialect) quoter() byte {
	return '`'
}
func (m mysqlDialect) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
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

type postgresDialect struct {
	standardSQL
}
