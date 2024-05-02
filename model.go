package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	tagKeyColumn = "column"
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	colName string
}

//var defaultRegistry = &registry{
//	models: make(map[reflect.Type]*model),
//}

// registry 元数据注册中心
type registry struct {
	models sync.Map
	//lock   sync.RWMutex
	//models map[reflect.Type]*model
}

func newRegistry() *registry {
	return &registry{}
}

func (r *registry) get(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)

	m, ok := r.models.Load(typ)

	if ok {
		return m.(*model), nil
	}

	var err error
	m, err = r.parseModel(entity)
	if err != nil {
		return nil, err
	}
	r.models.Store(typ, m)
	return m.(*model), nil
}

//func (r *registry) get1(entity any) (*model, error) {
//	typ := reflect.TypeOf(entity)
//	r.lock.RLock()
//	m, ok := r.models[typ]
//	r.lock.RUnlock()
//	if ok {
//		return m, nil
//	}
//	r.lock.Lock()
//	defer r.lock.Unlock()
//
//	m, ok = r.models[typ]
//	if ok {
//		return m, nil
//	}
//	var err error
//	m, err = r.parseModel(entity)
//	if err != nil {
//		return nil, err
//	}
//	r.models[typ] = m
//	return m, nil
//}

// parseModel 解析实体类限制只能用一级指针
func (r *registry) parseModel(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numField := typ.NumField()
	fieldMap := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		pairTag, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pairTag[tagKeyColumn]
		if colName == "" {
			// 如果没设置column
			colName = underscoreName(fd.Name)
		}
		fieldMap[fd.Name] = &field{
			colName: underscoreName(colName),
		}
	}
	var tableName string
	if tbname, ok := entity.(TableName); ok {
		tableName = tbname.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}
	return &model{
		tableName: tableName,
		fields:    fieldMap,
	}, nil
}

type User struct {
	ID uint64 `orm:"column"`
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok {
		return map[string]string{}, nil
	}
	pairs := strings.Split(ormTag, ",")
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		segs := strings.Split(pair, ":")
		if len(segs) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		key := segs[0]
		val := segs[1]
		res[key] = val
	}
	return res, nil
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
