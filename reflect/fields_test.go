package reflect

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type User struct {
	Name string
	age  int
}

func TestIterateField(t *testing.T) {
	testCases := []struct {
		name    string
		entity  any
		wantRes map[string]any
		wantErr error
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				// 忽略 age
				"age": 0,
			},
			wantErr: nil,
		},
		{
			name: "pointer",
			entity: &User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				// 忽略 age
				"age": 0,
			},
			wantErr: nil,
		},
		{
			name:    "basic type",
			entity:  13,
			wantRes: nil,
			wantErr: errors.New("entity must be a struct"),
		},
		{
			name: "multi pointer",
			entity: func() **User {
				res := &User{
					Name: "Tom",
					age:  18,
				}
				return &res
			}(),
			wantRes: map[string]any{
				"Name": "Tom",
				// 忽略 age
				"age": 0,
			},
			wantErr: nil,
		},
		{
			name:    "nil",
			entity:  nil,
			wantRes: nil,
			wantErr: errors.New("entity is nil"),
		},
		{
			name:    "user nil",
			entity:  (*User)(nil),
			wantRes: nil,
			wantErr: errors.New("entity is zero"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateField(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestSetField(t *testing.T) {
	testCases := []struct {
		name   string
		entity any
		field  string
		newVal any

		wantErr    error
		wantEntity any
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
			},
			field:  "Name",
			newVal: "Jerry",
			wantEntity: User{
				Name: "Jerry",
			},
			wantErr: errors.New("field can not set"),
		},
		{
			name: "pointer",
			entity: &User{
				Name: "Tom",
			},
			field:  "Name",
			newVal: "Jerry",
			wantEntity: &User{
				Name: "Jerry",
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetField(tc.entity, tc.field, tc.newVal)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantEntity, tc.entity)

		})
	}
}
