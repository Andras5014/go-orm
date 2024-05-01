package reflect

import (
	"github.com/Andras5014/go-orm/reflect/types"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestIterateFunc(t *testing.T) {
	testCases := []struct {
		name    string
		entity  any
		wantRes map[string]FuncInfo
		wantErr error
	}{
		{
			name:   "struct",
			entity: types.NewUser("Andras", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": {
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
					Name:        "GetAge",
					// 下表0指向接收器
					InputTypes: []reflect.Type{reflect.TypeOf(types.User{})},
				},
				//"ChangeName": {
				//	//OutputTypes: []reflect.Type{},
				//	//Result:      []any{},
				//	Name:       "ChangeName",
				//	InputTypes: []reflect.Type{reflect.TypeOf("")},
				//},
			},
		},
		{
			name:   "pointer",
			entity: types.NewUserPtr("Andras", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": {
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
					Name:        "GetAge",
					// 下表0指向接收器
					InputTypes: []reflect.Type{reflect.TypeOf(&types.User{})},
				},
				"ChangeName": {
					OutputTypes: []reflect.Type{},
					Result:      []any{},
					Name:        "ChangeName",
					InputTypes:  []reflect.Type{reflect.TypeOf(&types.User{}), reflect.TypeOf("")},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateFunc(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
