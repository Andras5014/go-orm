package go_orm

import (
	"database/sql"
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
				SQL:  "SELECT * FROM `TestModel`;",
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
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: (&Selector[TestModel]{}).Where(C("Name").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `Name` = ?;",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "not",
			builder: (&Selector[TestModel]{}).Where(Not(C("Name").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`Name` = ?);",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "and",
			builder: (&Selector[TestModel]{}).Where(C("Name").Eq("Tom").And(C("Id").Eq(123))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Name` = ?) AND (`Id` = ?);",
				Args: []any{"Tom", 123},
			},
		},
		{
			name:    "or",
			builder: (&Selector[TestModel]{}).Where(C("Name").Eq("Tom").Or(C("Id").Eq(123))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Name` = ?) OR (`Id` = ?);",
				Args: []any{"Tom", 123},
			},
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
