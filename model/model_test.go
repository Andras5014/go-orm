package model

import (
	"database/sql"
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestRegistry_Register(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		fields    []*Field
		wantErr   error
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Typ:     reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
		},
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "map",
			entity:  map[string]int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "array",
			entity:  [...]int{},
			wantErr: errs.ErrPointerOnly,
		},
	}
	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.entity)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range tc.fields {
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.EqualValues(t, tc.wantModel, m)

		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		fields    []*Field
		wantErr   error
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Typ:     reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
		},
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column:first_name_t"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name_t",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name: "empty column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column:"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name: "column only",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			name: "ignore tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"123:123"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
	}
	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range tc.fields {
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.EqualValues(t, tc.wantModel, m)
			//assert.Equal(t, tc.cacheSize, r.models.Range)

			typ := reflect.TypeOf(tc.entity)
			//model, ok := r.models.Load(Typ)
			model, ok := r.(*registry).models.Load(typ)
			assert.True(t, ok)
			assert.EqualValues(t, tc.wantModel, model)
		})
	}
}

type CustomTableName struct {
	FirstName string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	FirstName string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr"
}

func TestModelWithColumnName(t *testing.T) {

	testCases := []struct {
		name        string
		entity      any
		field       string
		colName     string
		wantErr     error
		wantColName string
		opts        []Option
	}{
		{
			name:        "column name",
			entity:      &TestModel{},
			field:       "FirstName",
			colName:     "first_name_t",
			wantColName: "first_name_t",
		},
		{
			name:    "invalid column name",
			entity:  &TestModel{},
			field:   "xxx",
			colName: "first_name_t",
			wantErr: errs.NewErrUnknownField("xxx"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			m, err := r.Register(tc.entity, WithColumnName(tc.field, tc.colName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd, ok := m.FieldMap[tc.field]
			require.True(t, ok)
			assert.Equal(t, tc.wantColName, fd.ColName)
			//assert.Equal(t, tc.cacheSize, r.models.Range)
		})
	}
}

func TestModelWithTableName(t *testing.T) {
	r := NewRegistry()
	m, err := r.Register(&TestModel{}, WithTableName("test_model_t"))
	require.NoError(t, err)
	assert.Equal(t, "test_model_t", m.TableName)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
