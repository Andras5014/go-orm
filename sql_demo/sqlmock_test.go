package sql_demo

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestSQLMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			return
		}
	}(db)
	require.NoError(t, err)
	mockRows := sqlmock.NewRows([]string{"id", "first_name"})
	mockRows.AddRow(1, "tom")
	mock.ExpectQuery("SELECT id, first_name FROM `user`.*").WillReturnRows(mockRows)
	rows, err := db.QueryContext(context.Background(), "SELECT id, first_name FROM `user` WHERE id=1")
	require.NoError(t, err)
	for rows.Next() {
		tm := TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName)
		require.NoError(t, err)
		log.Println(tm)

	}
}
