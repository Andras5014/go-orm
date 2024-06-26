package valuer

import (
	"database/sql"
	go_orm "github.com/Andras5014/go-orm/model"
)

type Value interface {
	Field(fd string) (any, error)
	SetColumns(rows *sql.Rows) error
}

type Creator func(model *go_orm.Model, entity any) Value
