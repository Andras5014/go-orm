package go_orm

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/valuer"
	"github.com/Andras5014/go-orm/model"
)

type DBOption func(db *DB)

// DB is the sqlDB wrapper
type DB struct {
	r       model.Registry
	db      *sql.DB
	creator valuer.Creator
}

func Open(driver string, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}

	return OpenDB(db, opts...)
}
func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:       model.NewRegistry(),
		db:      db,
		creator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}
func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}
func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}
func MustOpenDB(driver string, dataSourceName string, opts ...DBOption) *DB {
	db, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return db
}
