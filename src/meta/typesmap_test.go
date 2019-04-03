package meta

import "testing"

func BenchmarkTypesMapIterate(b *testing.B) {
	tests := []struct {
		name       string
		typeString string
	}{
		{"0", ""},
		{"1", "int"},
		{"2", "int|string"},
		{"3", "int|string|double"},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			m := NewTypesMap(test.typeString)
			for i := 0; i < b.N; i++ {
				m.Iterate(func(typ string) {})
			}
		})
	}
}
