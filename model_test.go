package go_orm

import (
	"github.com/Andras5014/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotModel, err := parseModel(tc.entity)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, gotModel)
		})
	}
}
