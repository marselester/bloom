// Package bloom provides a Bloom filter (space-efficient probabilistic data structure)
// that is used to test whether an element is a member of a set.
// False positive matches are possible, but false negatives are not – in other words, a query returns either "possibly in set" or
// "definitely not in set". Elements can be added to the set, but not removed; the more elements that are added to the set,
// the larger the probability of false positives.
//
// A Bloom filter of a fixed size can represent a set with an arbitrarily large number of elements
// (4,294,967,295 in this implementation); adding an element never fails due to the data structure "filling up".
package bloom

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strconv"
)

// Filter represents a Bloom filter.
// Note, operations are not concurrency safe.
type Filter struct {
	// prob is a desired probability of false positives.
	prob float64
	// bitlen is how many bits are needed to store n elements.
	bitlen uint64
	// hashqty is a number of hash functions.
	hashqty byte
	// n is a number of elements a client intends to store.
	n uint32
	// bitstore is a bit array of uint64 bit buckets.
	bitstore []uint64
}

// New creates a new Bloom filter for n elements based on
// tolerated error rate of false positives (whether set contains an element).
func New(n uint32, prob float64) (*Filter, error) {
	if n == 0 {
		return nil, ErrZeroElements
	}
	if prob <= 0 {
		return nil, ErrProbability
	}

	bf := Filter{
		n:    n,
		prob: prob,
	}
	bf.hashqty = optimalHashQty(bf.prob)
	bf.bitlen = optimalBitLen(n, bf.prob)
	// We use uint64 bit buckets to accommodate calculated bitlen.
	buckets := bf.bitlen / 64
	if bf.bitlen%64 != 0 {
		buckets++
	}
	bf.bitstore = make([]uint64, buckets)
	return &bf, nil
}

// Add adds an element to the set. The error in unlikely to happen,
// unless underlying hash function fails.
func (bf *Filter) Add(element []byte) error {
	pos, err := bitpositions(element, bf.hashqty, bf.bitlen)
	if err != nil {
		return err
	}

	var mask uint64
	for _, p := range pos {
		index, offset := bitlocation(p, 64)
		mask = 1 << offset
		bf.bitstore[index] |= mask
	}
	return nil
}

// Has tests if the element is in the set. The error in unlikely to happen,
// unless underlying hash function fails.
func (bf *Filter) Has(element []byte) (bool, error) {
	pos, err := bitpositions(element, bf.hashqty, bf.bitlen)
	if err != nil {
		return false, err
	}

	var mask uint64
	for _, p := range pos {
		index, offset := bitlocation(p, 64)
		mask = 1 << offset
		if (bf.bitstore[index] & mask) == 0 {
			return false, nil
		}
	}
	return true, nil
}

// MustAdd is similar to Add, but it panics if the error is not nil.
// Underlying hash function is cause of an error.
func (bf *Filter) MustAdd(element []byte) {
	if err := bf.Add(element); err != nil {
		panic(err)
	}
}

// MustHave is similar to Has, but it panics if the error is not nil.
// Underlying hash function is cause of an error.
func (bf *Filter) MustHave(element []byte) bool {
	isIn, err := bf.Has(element)
	if err != nil {
		panic(err)
	}
	return isIn
}

// optimalBitLen finds the optimal length of a bit array
// based on n number of elements in a set and prob error rate (probability of false positives).
func optimalBitLen(n uint32, prob float64) uint64 {
	ln2 := math.Log(2)
	optLen := -float64(n) * math.Log(prob) / (ln2 * ln2)
	return uint64(math.Ceil(optLen))
}

// optimalHashQty finds the optimal count of hash functions based on desired probability of an error.
func optimalHashQty(prob float64) byte {
	optQty := -math.Log(prob) / math.Log(2)
	return byte(math.Ceil(optQty))
}

// bitpositions applies hashQty hash functions to an element to calculate its bit positions.
// They are used to add an element or test whether it is in the set.
func bitpositions(element []byte, hashqty byte, bitlen uint64) ([]uint64, error) {
	var err error
	// We'll concat element and hash index to obtain hashQty bit positions.
	b := make([]byte, len(element)+1)
	copy(b, element)

	pos := make([]uint64, hashqty)
	for i := byte(0); i < hashqty; i++ {
		b[len(element)] = i
		pos[i], err = hash(b, bitlen)
		if err != nil {
			break
		}
	}
	return pos, err
}

// hash returns a position in the bit array by hashing b.
// sha256(b) hexdigest is converted to a number which is "truncated" to fit into bitlen range.
func hash(b []byte, bitlen uint64) (uint64, error) {
	h := sha256.New()
	if _, err := h.Write(b); err != nil {
		return 0, err
	}

	hexdigest := fmt.Sprintf("%x", h.Sum(nil))
	// We use first 16 chars of the hex digest to create a number.
	// If we use more chars, then it overflows.
	i, err := strconv.ParseUint(hexdigest[:16], 16, 64)
	if err != nil {
		return 0, err
	}
	// Fit i into the range of the bit array.
	return i % bitlen, nil
}

// bitlocation returns index in a bitstore and bit offset in bit bucket.
// If bitsize is zero, then bucket size is assumed to be 8 bits.
func bitlocation(p uint64, bitsize byte) (int, byte) {
	if bitsize == 0 {
		bitsize = 8
	}
	// index in a bitstore (round up)
	index := p / uint64(bitsize)
	// bit position in the bucket.
	offset := p - index*uint64(bitsize)
	return int(index), byte(offset)
}
