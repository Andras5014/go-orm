package querylog

import (
	"context"
	"database/sql"
	go_orm "github.com/Andras5014/go-orm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	var query string
	var args []any
	m := (&MiddlewareBuilder{}).LogFunc(func(q string, as []any) {
		query = q
		args = as

	})
	db, err := go_orm.Open("sqlite3", "file:test.db?cache=shared&mode=memory", go_orm.DBWithMiddlewares(m.Build()))
	require.NoError(t, err)
	_, _ = go_orm.NewSelector[TestModel](db).Where(go_orm.C("Id").Eq(1)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{1}, args)

	_ = go_orm.NewInserter[TestModel](db).Values(&TestModel{Id: 1, LastName: &sql.NullString{String: "test", Valid: true}}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model` (`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);", query)
	assert.EqualValues(t, []any{1, "", 0, &sql.NullString{String: "test", Valid: true}}, args)
}

type TestModel struct {
	Id        int
	FirstName string
	Age       int
	LastName  *sql.NullString
}
