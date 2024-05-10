package valuer

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	go_orm "github.com/Andras5014/go-orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	model *go_orm.Model
	// 起始地址
	address unsafe.Pointer
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *go_orm.Model, val any) Value {
	address := reflect.ValueOf(val).UnsafePointer()
	return unsafeValue{
		model:   model,
		address: address,
	}
}
func (u unsafeValue) Field(name string) (any, error) {
	fd, ok := u.model.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownColumn(name)
	}
	fdAddress := unsafe.Pointer(uintptr(u.address) + fd.Offset)
	val := reflect.NewAt(fd.Typ, fdAddress)
	return val.Elem().Interface(), nil
}
func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	// 拿到 select 出来的列
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	var vals []any
	for _, c := range cs {
		fd, ok := u.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 计算字段地址
		// 起始地址+字段偏移量
		fdAddress := unsafe.Pointer(uintptr(u.address) + fd.Offset)
		val := reflect.NewAt(fd.Typ, fdAddress)
		vals = append(vals, val.Interface())
	}
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}
	return err
}
