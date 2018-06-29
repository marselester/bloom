package bloom

import "testing"

func BenchmarkFilter_Add(b *testing.B) {
	tt := []struct {
		name string
		n    uint32
		prob float64
	}{
		{"1.198MB", 1000000, 0.01},
		{"2.573GB", 2147483647, 0.01},
	}

	for _, tc := range tt {
		b.Run(tc.name, func(b *testing.B) {
			bf, err := New(tc.n, tc.prob)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bf.Add([]byte("Hello, 世界 🤪"))
			}
		})
	}
}

func BenchmarkFilter_Has(b *testing.B) {
	tt := []struct {
		name string
		n    uint32
		prob float64
	}{
		{"1.198MB", 1000000, 0.01},
		{"2.573GB", 2147483647, 0.01},
	}

	for _, tc := range tt {
		b.Run(tc.name, func(b *testing.B) {
			bf, err := New(tc.n, tc.prob)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bf.Has([]byte("Hello, 世界 🤪"))
			}
		})
	}
}
