package model

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
	Register(entity any, opts ...Option) (*Model, error)
}
type Model struct {
	TableName string
	Fields    []*Field
	// 字段名 -> 字段
	FieldMap map[string]*Field
	// 列名 -> 字段
	ColumnMap map[string]*Field
}

type Option func(model *Model) error

type Field struct {
	ColName string
	//代码中字段名
	GoName string
	// 字段类型
	Typ reflect.Type

	// 字段相对于结构体偏移量
	Offset uintptr
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

func NewRegistry() Registry {
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
//	Typ := reflect.TypeOf(entity)
//	r.lock.RLock()
//	m, ok := r.models[Typ]
//	r.lock.RUnlock()
//	if ok {
//		return m, nil
//	}
//	r.lock.Lock()
//	defer r.lock.Unlock()
//
//	m, ok = r.models[Typ]
//	if ok {
//		return m, nil
//	}
//	var err error
//	m, err = r.Register(entity)
//	if err != nil {
//		return nil, err
//	}
//	r.models[Typ] = m
//	return m, nil
//}

// Register 解析实体类限制只能用一级指针
func (r *registry) Register(entity any, opts ...Option) (*Model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemTyp := typ.Elem()
	numField := elemTyp.NumField()
	fieldMap := make(map[string]*Field, numField)
	columnMap := make(map[string]*Field, numField)
	fields := make([]*Field, 0, numField)
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
			ColName: colName,
			Typ:     fd.Type,
			GoName:  fd.Name,
			Offset:  fd.Offset,
		}
		fieldMap[fd.Name] = fdMeta
		columnMap[colName] = fdMeta
		fields = append(fields, fdMeta)
	}
	var tableName string
	if tbname, ok := entity.(TableName); ok {
		tableName = tbname.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemTyp.Name())
	}

	res := &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
		Fields:    fields,
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

func WithTableName(tableName string) Option {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}
func WithColumnName(field string, colName string) Option {
	return func(model *Model) error {
		fd, ok := model.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		fd.ColName = colName
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

type TableName interface {
	TableName() string
}
