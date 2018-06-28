//+build gofuzz

package bloom

func Fuzz(data []byte) int {
	bf, _ := New(4294967295, 0.0001)
	if err := bf.Add(data); err != nil {
		return 0
	}
	return 1
}
