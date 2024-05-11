package go_orm

import (
	"context"
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/Andras5014/go-orm/internal/valuer"
	"github.com/Andras5014/go-orm/model"
)

type DBOption func(db *DB)

// DB is the sqlDB wrapper
type DB struct {
	core
	db *sql.DB
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
		core: core{
			r:       model.NewRegistry(),
			dialect: DialectMySQL,
			creator: valuer.NewUnsafeValue,
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
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

func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{
		tx: tx,
	}, nil
}

type txKey struct{}

// BeginTxV2 事务扩散
func (d *DB) BeginTxV2(ctx context.Context, opts *sql.TxOptions) (context.Context, *Tx, error) {
	val := ctx.Value(txKey{})
	tx, ok := val.(*Tx)
	if ok && !tx.done {
		return ctx, tx, nil
	}
	tx, err := d.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	return context.WithValue(ctx, txKey{}, tx), tx, nil
}

// BeginTxV3 要求前面一定要开启事务
//
//	func (d *DB) BeginTxV3(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
//		val := ctx.Value(txKey{})
//		tx, ok := val.(*Tx)
//		if ok {
//			return tx, nil
//		}
//
//		return nil, errors.New("no tx ")
//	}
func (d *DB) DoTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error, opts *sql.TxOptions) (err error) {
	tx, err := d.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	panicked := true

	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollback(err, e, panicked)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return nil
}
func (d *DB) getCore() core {
	return d.core
}
func (d *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}
