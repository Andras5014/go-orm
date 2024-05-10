package go_orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)
	//assert.NoError(t, err)
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
		{
			name:    "raw expression as predicate",
			builder: NewSelector[TestModel](db).Where(Raw("`id`<?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id`<?;",
				Args: []any{18},
			},
		},
		{
			name:    "raw expression used in predicate",
			builder: NewSelector[TestModel](db).Where(C("Id").Eq(Raw("`age`+?", 18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = `age`+?;",
				Args: []any{18},
			},
		},
		{
			name:    "columns alias",
			builder: NewSelector[TestModel](db).Select(C("FirstName").As("my_name")),
			wantQuery: &Query{
				SQL: "SELECT `first_name` AS `my_name` FROM `test_model`;",
			},
		},
		{
			name:    "avg alias",
			builder: NewSelector[TestModel](db).Select(Avg("FirstName").As("my_name")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`first_name`) AS `my_name` FROM `test_model`;",
			},
		},
		{
			name:    "columns alias in where",
			builder: NewSelector[TestModel](db).Where(C("FirstName").As("my_name").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `first_name` = ?;",
				Args: []any{"Tom"},
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

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	//对应query error
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))
	//对应 no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .* WHERE `id` < .*").WillReturnRows(rows)
	// data
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("123", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .* ").WillReturnRows(rows)

	testCases := []struct {
		name    string
		s       *Selector[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "invalid query",
			s:       NewSelector[TestModel](db).Where(C("XXXX").Eq("Tom")),
			wantErr: errs.NewErrUnknownField("XXXX"),
		},
		{
			name:    "query error",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(123)),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       NewSelector[TestModel](db).Where(C("Id").Lt(123)),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq(123)),
			wantRes: &TestModel{
				Id:        123,
				FirstName: "Tom",
				Age:       18,
				LastName: &sql.NullString{
					String: "Jerry",
					Valid:  true,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func memoryDB(t *testing.T, opts ...DBOption) *DB {
	db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory", opts...)
	require.NoError(t, err)
	return db
}

func TestSelector_Select(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "select",
			builder: NewSelector[TestModel](db).Select(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "SELECT `first_name`,`last_name` FROM `test_model`;",
			},
		},
		{
			name:    "select all",
			builder: NewSelector[TestModel](db).Select(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    " invalid",
			builder: NewSelector[TestModel](db).Select(C("invalid")),
			wantErr: errs.NewErrUnknownField("invalid"),
		},

		{
			name:    " avg",
			builder: NewSelector[TestModel](db).Select(Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name:    " max",
			builder: NewSelector[TestModel](db).Select(Max("Age")),
			wantQuery: &Query{
				SQL: "SELECT MAX(`age`) FROM `test_model`;",
			},
		},
		{
			name:    " min",
			builder: NewSelector[TestModel](db).Select(Min("Age")),
			wantQuery: &Query{
				SQL: "SELECT MIN(`age`) FROM `test_model`;",
			},
		},
		{
			name:    " count",
			builder: NewSelector[TestModel](db).Select(Count("Age")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(`age`) FROM `test_model`;",
			},
		},
		{
			name:    " sum",
			builder: NewSelector[TestModel](db).Select(Sum("Age")),
			wantQuery: &Query{
				SQL: "SELECT SUM(`age`) FROM `test_model`;",
			},
		},
		{
			name:    "min invalid columns",
			builder: NewSelector[TestModel](db).Select(Min("invalid")),
			wantErr: errs.NewErrUnknownField("invalid"),
		},
		{
			name:    "max invalid columns",
			builder: NewSelector[TestModel](db).Select(Max("invalid")),
			wantErr: errs.NewErrUnknownField("invalid"),
		},
		{
			name:    "count invalid columns",
			builder: NewSelector[TestModel](db).Select(Count("invalid")),
			wantErr: errs.NewErrUnknownField("invalid"),
		},
		{
			name:    "multiple aggregate",
			builder: NewSelector[TestModel](db).Select(Avg("Age"), Max("Age"), Min("Age"), Count("Age"), Sum("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`),MAX(`age`),MIN(`age`),COUNT(`age`),SUM(`age`) FROM `test_model`;",
			},
		},
		{
			name:    "raw expression",
			builder: NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT `first_name`)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT `first_name`) FROM `test_model`;",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.EqualValues(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
