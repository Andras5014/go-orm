package integration

import (
	go_orm "github.com/Andras5014/go-orm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	driver string
	dsn    string
	db     *go_orm.DB
}

func (s *Suite) SetupSuite() {
	db, err := go_orm.Open(s.driver, s.dsn)
	require.NoError(s.T(), err)
	s.db = db
}
