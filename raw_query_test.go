package go_orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawQuery_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	//对应query error
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))
	//对应 no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .* ").WillReturnRows(rows)
	// data
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .* ").WillReturnRows(rows)

	testCases := []struct {
		name    string
		s       *RawQuerier[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "query error",
			s:       RawQuery[TestModel](db, "SELECT * FORM `test_model`"),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       RawQuery[TestModel](db, "SELECT * FORM `test_model` WHERE `id` = ?", -1),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			s:    RawQuery[TestModel](db, "SELECT * FORM `test_model` WHERE `id` = ?", 1),
			wantRes: &TestModel{
				Id:        1,
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
