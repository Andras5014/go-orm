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

type Registry interface {
	Get(entity any) (*Model, error)
	Register(entity any, opts ...ModelOption) (*Model, error)
}
type Model struct {
	tableName string
	// 字段名 -> 字段
	fieldMap map[string]*Field
	// 列名 -> 字段
	columnMap map[string]*Field
}

type ModelOption func(model *Model) error

type Field struct {
	colName string
	//代码中字段名
	goName string
	// 字段类型
	typ reflect.Type
}

//var defaultRegistry = &registry{
//	models: make(map[reflect.Type]*Model),
//}

// registry 元数据注册中心
type registry struct {
	models sync.Map
	//lock   sync.RWMutex
	//models map[reflect.Type]*Model
}

func newRegistry() *registry {
	return &registry{}
}

func (r *registry) Get(entity any) (*Model, error) {
	typ := reflect.TypeOf(entity)

	m, ok := r.models.Load(typ)

	if ok {
		return m.(*Model), nil
	}

	var err error
	m, err = r.Register(entity)
	if err != nil {
		return nil, err
	}

	return m.(*Model), nil
}

//func (r *registry) get1(entity any) (*Model, error) {
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
//	m, err = r.Register(entity)
//	if err != nil {
//		return nil, err
//	}
//	r.models[typ] = m
//	return m, nil
//}

// Register 解析实体类限制只能用一级指针
func (r *registry) Register(entity any, opts ...ModelOption) (*Model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemTyp := typ.Elem()
	numField := elemTyp.NumField()
	fieldMap := make(map[string]*Field, numField)
	columnMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fd := elemTyp.Field(i)
		pairTag, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pairTag[tagKeyColumn]
		if colName == "" {
			// 如果没设置column
			colName = underscoreName(fd.Name)
		}
		fdMeta := &Field{
			colName: colName,
			typ:     fd.Type,
			goName:  fd.Name,
		}
		fieldMap[fd.Name] = fdMeta
		columnMap[colName] = fdMeta
	}
	var tableName string
	if tbname, ok := entity.(TableName); ok {
		tableName = tbname.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemTyp.Name())
	}

	res := &Model{
		tableName: tableName,
		fieldMap:  fieldMap,
		columnMap: columnMap,
	}
	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}
	r.models.Store(typ, res)
	return res, nil
}

func ModelWithTableName(tableName string) ModelOption {
	return func(model *Model) error {
		model.tableName = tableName
		return nil
	}
}
func ModelWithColumnName(field string, colName string) ModelOption {
	return func(model *Model) error {
		fd, ok := model.fieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		fd.colName = colName
		return nil
	}
}

//type User struct {
//	ID uint64 `orm:"column"`
//}

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
