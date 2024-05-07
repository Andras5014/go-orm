package unsafe

import "testing"

func TestPrintFieldOffset(t *testing.T) {

	testCases := []struct {
		name   string
		entity any
	}{
		{
			name:   "user",
			entity: User{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			PrintFieldOffset(tc.entity)
		})
	}
}

type User struct {
	Name    string
	Age     int8
	abc     int8
	Alias   []string
	Address string
}
