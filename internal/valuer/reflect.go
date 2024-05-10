package valuer

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	go_orm "github.com/Andras5014/go-orm/model"
	"reflect"
)

type reflectValue struct {
	model *go_orm.Model
	// T的指针
	val reflect.Value
}

var _ Creator = NewReflectValue

func NewReflectValue(model *go_orm.Model, val any) Value {
	return reflectValue{
		val:   reflect.ValueOf(val),
		model: model,
	}
}
func (r reflectValue) Field(name string) (any, error) {
	// 检测name是否合法
	//_, ok := r.val.Type().FieldByName(name)
	//if !ok {
	//	return nil, errs.NewErrUnknownField(name)
	//}

	return r.val.FieldByName(name).Interface(), nil
}
func (r reflectValue) SetColumns(rows *sql.Rows) error {
	// 拿到 select 出来的列
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		val := reflect.New(fd.Typ)
		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())

	}
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}
	tpValueElem := r.val
	for i, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		tpValueElem.FieldByName(fd.GoName).Set(valElems[i])

	}
	return nil
}
