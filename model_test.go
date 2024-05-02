package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_parseModel(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *model
		wantErr   error
	}{
		{
			name:   "pointer",
			entity: TestModel{},
			wantModel: &model{
				tableName: "test_model",
				fields: map[string]*field{
					"Id":        &field{colName: "id"},
					"FirstName": &field{colName: "first_name"},
					"LastName":  &field{colName: "last_name"},
					"Age":       &field{colName: "age"},
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
	r := newRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotModel, err := r.parseModel(tc.entity)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, gotModel)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *model
		wantErr   error
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &model{
				tableName: "test_model",
				fields: map[string]*field{
					"Id":        &field{colName: "id"},
					"FirstName": &field{colName: "first_name"},
					"LastName":  &field{colName: "last_name"},
					"Age":       &field{colName: "age"},
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
			wantModel: &model{
				tableName: "tag_table",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name_t",
					},
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
			wantModel: &model{
				tableName: "tag_table",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
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
			wantModel: &model{
				tableName: "tag_table",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &model{
				tableName: "custom_table_name_t",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &model{
				tableName: "custom_table_name_ptr",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
	}
	r := newRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
			//assert.Equal(t, tc.cacheSize, r.models.Range)

			typ := reflect.TypeOf(tc.entity)
			model, ok := r.models.Load(typ)
			assert.True(t, ok)
			assert.Equal(t, tc.wantModel, model)
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
