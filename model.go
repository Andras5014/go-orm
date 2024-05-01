package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"reflect"
	"unicode"
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	colName string
}

// parseModel 解析实体类限制只能用一级指针
func parseModel(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numField := typ.NumField()
	fieldMap := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fieldMap[fd.Name] = &field{
			colName: underscoreName(fd.Name),
		}
	}
	return &model{
		tableName: underscoreName(typ.Name()),
		fields:    fieldMap,
	}, nil
}

// underscoreName 将驼峰命名转换为下划线命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, V := range tableName {
		if unicode.IsUpper(V) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(V)))
		} else {
			buf = append(buf, byte(V))
		}
	}
	return string(buf)
}
