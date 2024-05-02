package go_orm

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db, err := NewDB()
	assert.NoError(t, err)
	testCases := []struct {
		name string

		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "from",
			builder: NewSelector[TestModel](db).From("`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: NewSelector[TestModel](db).Where(C("FirstName").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `first_name` = ?;",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "not",
			builder: NewSelector[TestModel](db).Where(Not(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`first_name` = ?);",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "and",
			builder: NewSelector[TestModel](db).Where(C("FirstName").Eq("Tom").And(C("Id").Eq(123))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`first_name` = ?) AND (`id` = ?);",
				Args: []any{"Tom", 123},
			},
		},
		{
			name:    "or",
			builder: NewSelector[TestModel](db).Where(C("FirstName").Eq("Tom").Or(C("Id").Eq(123))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`first_name` = ?) OR (`id` = ?);",
				Args: []any{"Tom", 123},
			},
		},
		{
			name:    "invalid column",
			builder: NewSelector[TestModel](db).Where(C("XXXX").Eq("Tom")),
			wantErr: errs.NewErrUnknownField("XXXX"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantQuery, q)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
