package go_orm

import (
	"github.com/Andras5014/go-orm/internal/valuer"
	"github.com/Andras5014/go-orm/model"
)

type core struct {
	model       *model.Model
	dialect     Dialect
	creator     valuer.Creator
	r           model.Registry
	middlewares []Middleware
}
