package go_orm

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {

	testCases := []struct {
		name string

		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: &Selector[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "from",
			builder: (&Selector[TestModel]{}).From("`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: (&Selector[TestModel]{}).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: (&Selector[TestModel]{}).Where(C("FirstName").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `first_name` = ?;",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "not",
			builder: (&Selector[TestModel]{}).Where(Not(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`first_name` = ?);",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "and",
			builder: (&Selector[TestModel]{}).Where(C("FirstName").Eq("Tom").And(C("Id").Eq(123))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`first_name` = ?) AND (`id` = ?);",
				Args: []any{"Tom", 123},
			},
		},
		{
			name:    "or",
			builder: (&Selector[TestModel]{}).Where(C("FirstName").Eq("Tom").Or(C("Id").Eq(123))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`first_name` = ?) OR (`id` = ?);",
				Args: []any{"Tom", 123},
			},
		},
		{
			name:    "invalid column",
			builder: (&Selector[TestModel]{}).Where(C("XXXX").Eq("Tom")),
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
