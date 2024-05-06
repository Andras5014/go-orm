package sql_demo

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

var createTableSql = `
CREATE TABLE IF NOT EXISTS test_model (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	first_name TEXT NOT NULL,
    	last_name TEXT NOT NULL
);
`

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()

	db.Ping()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	_, err = db.ExecContext(ctx, createTableSql)
	defer cancel()
	require.NoError(t, err)

	res, err := db.ExecContext(context.Background(), "INSERT INTO test_model (`id`, `first_name`,`last_name`) VALUES (?, ? , ?)", 1, "tom", "jerry")
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println(affected)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println(id)

	row := db.QueryRowContext(ctx, "SELECT * FROM test_model WHERE `id` = ?", 1)
	require.NoError(t, row.Err())
	tm := &TestModel{}
	row.Scan(&tm.Id, &tm.FirstName, &tm.LastName)
	log.Println(tm)
}

type TestModel struct {
	Id        int64
	FirstName string
	LastName  *sql.NullString
}
