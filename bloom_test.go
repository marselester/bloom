package bloom

import (
	"fmt"
	"testing"
)

func TestOptimalBitLen(t *testing.T) {
	tt := []struct {
		n    uint32
		prob float64
		want uint64
	}{
		{1000000, 0.01, 9585059}, // 1.198 MB
		{0, 0.01, 0},
		{2147483647, 0.01, 20583756121}, // 2.573 GB
		{4294967295, 0.01, 41167512252}, // 5.146 GB
	}

	for _, tc := range tt {
		got := optimalBitLen(tc.n, tc.prob)
		if got != tc.want {
			t.Errorf("optimalBitLen(%d, %f) = %d, want %d", tc.n, tc.prob, got, tc.want)
		}
	}
}

func TestOptimalHashQty(t *testing.T) {
	tt := []struct {
		prob float64
		want byte
	}{
		{0.1, 4},
		{0.01, 7},
		{0.0123, 7},
		{0.001, 10},
		{0.0001001231231, 14},
	}

	for _, tc := range tt {
		got := optimalHashQty(tc.prob)
		if got != tc.want {
			t.Errorf("optimalHashQty(%f) = %d, want %d", tc.prob, got, tc.want)
		}
	}
}

func TestHash(t *testing.T) {
	tt := []struct {
		b      string
		bitlen uint64
		want   uint64
	}{
		{"test", 1000000, 842533},
		{"test", 18446744073709551615, 11495104353665842533},
		{"test0", 48, 8},
		{"test1", 48, 24},
		{"test2", 48, 17},
		{"test3", 48, 23},
	}

	for _, tc := range tt {
		got, err := hash([]byte(tc.b), tc.bitlen)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("hash(%q, %d) = %d, want %d", tc.b, tc.bitlen, got, tc.want)
		}
	}
}

func equal(s1, s2 []uint64) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func TestBitpositions(t *testing.T) {
	tt := []struct {
		element string
		hashqty byte
		bitlen  uint64
		want    []uint64
	}{
		{"test", 4, 48, []uint64{7, 36, 32, 37}},
	}

	for _, tc := range tt {
		got, err := bitpositions([]byte(tc.element), tc.hashqty, tc.bitlen)
		if err != nil {
			t.Fatal(err)
		}
		if !equal(got, tc.want) {
			t.Errorf("bitpositions(%q, %d, %d) = %v, want %v", tc.element, tc.hashqty, tc.bitlen, got, tc.want)
		}
	}
}

func TestBitlocation(t *testing.T) {
	tt := []struct {
		pos        uint64
		bitsize    byte
		wantIndex  int
		wantOffset byte
	}{
		{
			pos:        0,
			bitsize:    2,
			wantIndex:  0,
			wantOffset: 0,
		},
		{
			pos:        1,
			bitsize:    2,
			wantIndex:  0,
			wantOffset: 1,
		},
		{
			pos:        2,
			bitsize:    2,
			wantIndex:  1,
			wantOffset: 0,
		},
		{
			pos:        3,
			bitsize:    2,
			wantIndex:  1,
			wantOffset: 1,
		},
		{
			pos:        41167512252,
			bitsize:    2,
			wantIndex:  20583756126,
			wantOffset: 0,
		},
		// Default bitsize is 8.
		{
			pos:        0,
			bitsize:    0,
			wantIndex:  0,
			wantOffset: 0,
		},
		{
			pos:        7,
			bitsize:    0,
			wantIndex:  0,
			wantOffset: 7,
		},
		{
			pos:        8,
			bitsize:    0,
			wantIndex:  1,
			wantOffset: 0,
		},
		{
			pos:        63,
			bitsize:    64,
			wantIndex:  0,
			wantOffset: 63,
		},
		{
			pos:        64,
			bitsize:    64,
			wantIndex:  1,
			wantOffset: 0,
		},
	}

	for _, tc := range tt {
		index, offset := bitlocation(tc.pos, tc.bitsize)
		if index != tc.wantIndex || offset != tc.wantOffset {
			t.Errorf("bitlocation(%d, %d) = %d, %d, want %d, %d", tc.pos, tc.bitsize, index, offset, tc.wantIndex, tc.wantOffset)
		}
	}
}

func TestFilter_Add(t *testing.T) {
	element := []byte("test")
	bf := &Filter{
		hashqty:  4,
		bitlen:   48,
		bitstore: make([]uint64, 1),
	}
	err := bf.Add(element)
	if err != nil {
		t.Fatal(err)
	}

	got := fmt.Sprintf("%064b", bf.bitstore[0])
	// bit positions: 7, 36, 32, 37
	want := "0000000000000000000000000011000100000000000000000000000010000000"
	if got != want {
		t.Errorf("Add(%q) %s, want %s", element, got, want)
	}
}

func TestFilter_Has(t *testing.T) {
	bf := &Filter{
		hashqty:  4,
		bitlen:   48,
		bitstore: []uint64{210453397632}, // "test" int representation of bit positions.
	}

	tt := []struct {
		element []byte
		want    bool
	}{
		{[]byte("test"), true},
		{[]byte("test1"), false},
		{nil, false},
	}

	for _, tc := range tt {
		got, err := bf.Has(tc.element)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("Has(%q) is %t, want %t", tc.element, got, tc.want)
		}
	}
}

func TestNew_error(t *testing.T) {
	tt := []struct {
		n    uint32
		prob float64
		want error
	}{
		{0, 0.1, ErrZeroElements},
		{1, 0, ErrProbability},
		{1, -0.1, ErrProbability},
	}

	for _, tc := range tt {
		_, err := New(tc.n, tc.prob)
		if err != tc.want {
			t.Errorf("New(%d, %f) error: %q, want %q", tc.n, tc.prob, err, tc.want)
		}
	}
}

func TestNew(t *testing.T) {
	tt := []struct {
		name string
		n    uint32
		prob float64
		want Filter
	}{
		{
			name: "n=1",
			n:    1,
			prob: 0.01,
			want: Filter{
				prob:     0.01,
				bitlen:   10,
				hashqty:  7,
				n:        1,
				bitstore: make([]uint64, 1),
			},
		},
		{
			name: "n=6",
			n:    6,
			prob: 0.01,
			want: Filter{
				prob:     0.01,
				bitlen:   58,
				hashqty:  7,
				n:        6,
				bitstore: make([]uint64, 1),
			},
		},
		{
			name: "n=7",
			n:    7,
			prob: 0.01,
			want: Filter{
				prob:     0.01,
				bitlen:   68,
				hashqty:  7,
				n:        7,
				bitstore: make([]uint64, 2),
			},
		},
		{
			name: "n=max",
			n:    4294967295,
			prob: 0.01,
			want: Filter{
				prob:     0.01,
				bitlen:   41167512252,
				hashqty:  7,
				n:        4294967295,
				bitstore: make([]uint64, 643242379),
			},
		},
		{
			name: "n=max prob=0.001",
			n:    4294967295,
			prob: 0.001,
			want: Filter{
				prob:     0.001,
				bitlen:   61751268378,
				hashqty:  10,
				n:        4294967295,
				bitstore: make([]uint64, 964863569),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := New(tc.n, tc.prob)
			if err != nil {
				t.Fatal(err)
			}

			if got.n != tc.want.n {
				t.Errorf("New(%d, %f) n = %d, want %d", tc.n, tc.prob, got.n, tc.want.n)
			}
			if got.prob != tc.want.prob {
				t.Errorf("New(%d, %f) prob = %f, want %f", tc.n, tc.prob, got.prob, tc.want.prob)
			}
			if got.bitlen != tc.want.bitlen {
				t.Errorf("New(%d, %f) bitlen = %d, want %d", tc.n, tc.prob, got.bitlen, tc.want.bitlen)
			}
			if got.hashqty != tc.want.hashqty {
				t.Errorf("New(%d, %f) hashqty = %d, want %d", tc.n, tc.prob, got.hashqty, tc.want.hashqty)
			}
			if len(got.bitstore) != len(tc.want.bitstore) {
				t.Errorf("New(%d, %f) bitstore len = %d, want %d", tc.n, tc.prob, len(got.bitstore), len(tc.want.bitstore))
			}
		})
	}
}
