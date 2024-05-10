package go_orm

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
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
