package go_orm

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
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
