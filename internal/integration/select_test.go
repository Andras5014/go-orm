package integration

import (
	"context"
	go_orm "github.com/Andras5014/go-orm"
	"github.com/Andras5014/go-orm/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SelectSuite struct {
	Suite
}

func TestMySQlSelect(t *testing.T) {
	suite.Run(t, &SelectSuite{
		Suite{
			dsn:    "root:root@tcp(localhost:3307)/integration_test",
			driver: "mysql",
		},
	})
}
func (s *SelectSuite) SetupSuite() {
	s.Suite.SetupSuite()
	res := go_orm.NewInserter[test.SimpleStruct](s.db).Values(test.NewSimpleStruct(41), test.NewSimpleStruct(42)).Exec(context.Background())
	require.NoError(s.T(), res.Err())
}
func (s *SelectSuite) TestSelect() {
	testCases := []struct {
		name    string
		s       *go_orm.Selector[test.SimpleStruct]
		wantRes *test.SimpleStruct
		wantErr error
	}{
		{
			name:    "get data",
			s:       go_orm.NewSelector[test.SimpleStruct](s.db).Where(go_orm.C("Id").Eq(41)),
			wantRes: test.NewSimpleStruct(41),
		},
		{
			name:    "no row",
			s:       go_orm.NewSelector[test.SimpleStruct](s.db).Where(go_orm.C("Id").Eq(100)),
			wantErr: go_orm.ErrNoRows,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			res, err := tc.s.Get(ctx)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
