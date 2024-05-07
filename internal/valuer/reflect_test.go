package valuer

import (
	"database/sql"
	go_orm "github.com/Andras5014/go-orm/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func TestNewReflectValue(t *testing.T) {
	testSetColumns(t, NewReflectValue)
}

func testSetColumns(t *testing.T, creator Creator) {
	testCases := []struct {
		name string
		// 一定是指针
		entity     any
		rows       *sqlmock.Rows
		wantErr    error
		wantEntity any
	}{
		{
			name:   "set columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow(1, "Andras", 18, "5014")
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Andras",
				Age:       18,
				LastName: &sql.NullString{
					String: "5014",
					Valid:  true,
				},
			},
		},
		{
			name:   "order",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
				rows.AddRow(1, "Andras", "5014", 18)
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Andras",
				Age:       18,
				LastName: &sql.NullString{
					String: "5014",
					Valid:  true,
				},
			},
		},
		{
			name:   "partial columns",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name"})
				rows.AddRow(1, "Andras")
				return rows
			}(),
			wantEntity: &TestModel{
				Id:        1,
				FirstName: "Andras",
			},
		},
	}
	r := go_orm.NewRegistry()
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = mockDB.Close()
	}()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构造rows
			mockRows := tc.rows
			mock.ExpectQuery("SELECT .*").WillReturnRows(mockRows)
			rows, err := mockDB.Query("SELECT .*")
			require.NoError(t, err)

			rows.Next()
			m, err := r.Register(tc.entity)
			require.NoError(t, err)
			val := creator(m, tc.entity)

			err = val.SetColumns(rows)
			require.NoError(t, err)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantEntity, tc.entity)
		})
	}
}
