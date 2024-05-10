package go_orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInserter_SQLite_upsert(t *testing.T) {
	db := memoryDB(t, DBWithDialect(DialectSQLite))
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name: "upsert-update value",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "a",
				Age:       18,
				LastName: &sql.NullString{
					String: "ndras",
					Valid:  true,
				},
			}).onDuplicateKey().ConflictColumns("FirstName", "Age").Update(Assign("FirstName", "J"), Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?)" +
					" ON CONFLICT (`first_name`,`age`) DO UPDATE SET `first_name`=?,`age`=?;",
				Args: []any{int64(1), "a", int8(18), &sql.NullString{String: "ndras", Valid: true}, "J", 19},
			},
		},
		{

			name: "upsert -update column",
			q: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        1,
				FirstName: "a",
			}, &TestModel{
				Id:        2,
				FirstName: "b",
			}).onDuplicateKey().ConflictColumns("FirstName", "Age").Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`) VALUES (?,?),(?,?)" +
					" ON CONFLICT (`first_name`,`age`) DO UPDATE SET `first_name`=EXCLUDED.`first_name`,`age`=EXCLUDED.`age`;",
				Args: []any{int64(1), "a",
					int64(2), "b"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
func TestInserter_Build(t *testing.T) {
	db := memoryDB(t, DBWithDialect(DialectMySQL))
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name: "single row",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "a",
				Age:       18,
				LastName: &sql.NullString{
					String: "ndras",
					Valid:  true,
				},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(1), "a", int8(18), &sql.NullString{String: "ndras", Valid: true}},
			},
		},
		{
			name: "multiple row",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "a",
				Age:       18,
				LastName: &sql.NullString{
					String: "ndras",
					Valid:  true,
				},
			}, &TestModel{
				Id:        2,
				FirstName: "b",
				Age:       28,
				LastName: &sql.NullString{
					String: "ndras",
					Valid:  true,
				},
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{int64(1), "a", int8(18), &sql.NullString{String: "ndras", Valid: true},
					int64(2), "b", int8(28), &sql.NullString{String: "ndras", Valid: true}},
			},
		},
		{
			name:    "no row",
			q:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			// 插入多行部分列
			name: "multiple row with partial columns",
			q: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        1,
				FirstName: "a",
			}, &TestModel{
				Id:        2,
				FirstName: "b",
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`) VALUES (?,?),(?,?);",
				Args: []any{int64(1), "a",
					int64(2), "b"},
			},
		},
		{
			name: "upsert-update value",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "a",
				Age:       18,
				LastName: &sql.NullString{
					String: "ndras",
					Valid:  true,
				},
			}).onDuplicateKey().Update(Assign("FirstName", "J"), Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?)" +
					" ON DUPLICATE KEY UPDATE `first_name`=?,`age`=?;",
				Args: []any{int64(1), "a", int8(18), &sql.NullString{String: "ndras", Valid: true}, "J", 19},
			},
		},
		{

			name: "upsert -update column",
			q: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        1,
				FirstName: "a",
			}, &TestModel{
				Id:        2,
				FirstName: "b",
			}).onDuplicateKey().Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`) VALUES (?,?),(?,?)" +
					" ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`age`=VALUES(`age`);",
				Args: []any{int64(1), "a",
					int64(2), "b"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)

		})
	}
}
func TestInserter_Exec(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		i        *Inserter[TestModel]
		wantErr  error
		affected int64
	}{
		{
			name: "db error",
			i: func() *Inserter[TestModel] {
				mock.ExpectExec("INSERT INTO .*").WillReturnError(errors.New("db error"))
				return NewInserter[TestModel](db).Values(&TestModel{
					Id:        1,
					FirstName: "a",
					Age:       18,
					LastName: &sql.NullString{
						String: "ndras",
						Valid:  true,
					},
				})
			}(),
			wantErr: errors.New("db error"),
		},
		{
			name: "query error",
			i: func() *Inserter[TestModel] {
				return NewInserter[TestModel](db).Values(&TestModel{}).Columns("Invalid")
			}(),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "exec",
			i: func() *Inserter[TestModel] {
				res := driver.RowsAffected(1)
				mock.ExpectExec("INSERT INTO .*").WillReturnResult(res)
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			affected: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, tc.wantErr, res.Err())
			if err != nil {
				return
			}
			assert.Equal(t, tc.affected, affected)

		})
	}
}
