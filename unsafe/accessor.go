package unsafe

import (
	"reflect"
	"unsafe"
)

type FieldMeta struct {
	Offset uintptr
	typ    reflect.Type
}

type UnsafeAccessor struct {
	fields  map[string]FieldMeta
	address unsafe.Pointer
}

func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	typ := reflect.TypeOf(entity).Elem()
	numFields := typ.NumField()
	fields := make(map[string]FieldMeta, numFields)
	for i := 0; i < numFields; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = FieldMeta{
			Offset: fd.Offset,
			typ:    fd.Type,
		}
	}
	val := reflect.ValueOf(entity)
	return &UnsafeAccessor{
		fields:  fields,
		address: val.UnsafePointer(),
	}
}

func (a *UnsafeAccessor) Field(field string) (any, error) {
	fieldAddress := unsafe.Pointer(uintptr(a.address) + a.fields[field].Offset)
	return reflect.NewAt(a.fields[field].typ, fieldAddress).Elem().Interface(), nil
}
func (a *UnsafeAccessor) SetField(field string, val any) error {
	fieldAddress := unsafe.Pointer(uintptr(a.address) + a.fields[field].Offset)
	reflect.NewAt(a.fields[field].typ, fieldAddress).Elem().Set(reflect.ValueOf(val))
	return nil

}
