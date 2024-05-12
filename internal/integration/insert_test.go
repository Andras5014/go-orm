package integration

import (
	"context"
	"database/sql"
	go_orm "github.com/Andras5014/go-orm"
	"github.com/Andras5014/go-orm/internal/test"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type InsertSuite struct {
	Suite
}

func TestMySQLInsert(t *testing.T) {
	suite.Run(t, &InsertSuite{
		Suite{
			dsn:    "root:root@tcp(localhost:3307)/integration_test",
			driver: "mysql",
		},
	})
}

func (i *InsertSuite) TestInsert() {
	t := i.T()

	testCases := []struct {
		name         string
		i            *go_orm.Inserter[test.SimpleStruct]
		wantAffected int64
	}{
		{
			name:         "insert one",
			i:            go_orm.NewInserter[test.SimpleStruct](i.db).Values(test.NewSimpleStruct(1)),
			wantAffected: 1,
		},
		{
			name: "insert multi",
			i: go_orm.NewInserter[test.SimpleStruct](i.db).Values(
				test.NewSimpleStruct(12),
				test.NewSimpleStruct(23)),
			wantAffected: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			res := tc.i.Exec(ctx)
			affected, err := res.RowsAffected()
			require.NoError(t, err)
			assert.Equal(t, tc.wantAffected, affected)

		})
	}
}

//	func TestMySQLInsert(t *testing.T) {
//		testInsert(t, "mysql", "root:root@tcp(localhost:3307)/integration_test")
//
// }
//
//	func testInsert(t *testing.T, driver string, dsn string) {
//		db, err := go_orm.Open(driver, dsn)
//		require.NoError(t, err)
//		testCases := []struct {
//			name         string
//			i            *go_orm.Inserter[test.SimpleStruct]
//			wantAffected int64
//		}{
//			{
//				name:         "insert one",
//				i:            go_orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(1)),
//				wantAffected: 1,
//			},
//			{
//				name: "insert multi",
//				i: go_orm.NewInserter[test.SimpleStruct](db).Values(
//					test.NewSimpleStruct(12),
//					test.NewSimpleStruct(23)),
//				wantAffected: 2,
//			},
//		}
//		for _, tc := range testCases {
//			t.Run(tc.name, func(t *testing.T) {
//				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//				defer cancel()
//				res := tc.i.Exec(ctx)
//				affected, err := res.RowsAffected()
//				require.NoError(t, err)
//				assert.Equal(t, tc.wantAffected, affected)
//
//			})
//		}
//	}
type SQLite3InsertSuite struct {
	InsertSuite
	driver string
	dsn    string
}

func (i *SQLite3InsertSuite) SetupSuite() {
	db, err := sql.Open(i.driver, i.dsn)
	// 建表
	_, err = db.ExecContext(context.Background(), "")

	require.NoError(i.T(), err)
	i.db, err = go_orm.OpenDB(db)
	require.NoError(i.T(), err)
}
func TestSQLite3Insert(t *testing.T) {
	suite.Run(t, &SQLite3InsertSuite{
		driver: "sqlite3",
		dsn:    "file:test.db?cache=shared&mode=memory",
	})
}
