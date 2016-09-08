// Package bloom package implement a simple and scalable Bloom Filter algorithm
package bloom

import (
	"encoding/json"
	"fmt"
	"math"
)

// Hasher : Pluggable hasher type
type Hasher func(string) uint64

// Filter : Implement a simple Filter
type Filter struct {
	arr      []byte
	Size     uint64
	k        uint64
	inserted uint64
	hasher   func([]byte) []uint64
}

// New : constructor
func New(size uint64, k uint64) *Filter {
	return &Filter{
		arr:      make([]byte, size),
		k:        k,
		Size:     size,
		inserted: 0,
		hasher:   generateHasher(k, size*8),
	}
}

// Reset : zeroes the bytearray, flushing the filter
func (bf *Filter) Reset() *Filter {
	bf.arr = make([]byte, bf.Size)
	return bf
}

// Match : Check if s have an entry in the filter
// May return false positive
func (bf *Filter) Match(s string) bool {
	hashs := bf.hasher([]byte(s))
	sectionSize := ((bf.Size * 8) / bf.k)
	for hid := uint64(0); hid < bf.k; hid++ {
		start := hid * sectionSize
		if !bf.isSet(start + (hashs[hid] % (sectionSize))) {
			return false
		}
	}
	return true
}

// ToJSON : Export a byte array that can be later used with bf.FromJSON
func (bf *Filter) ToJSON() ([]byte, error) {
	return json.Marshal(bf)
}

// FromJSON : Import a JSON serialized bloom filter
func FromJSON(json []byte) *Filter {
	// TODO
	return nil
}

// Merge two Filters, filters must have the same size
// Take care of fillratio when merging filters : false positive
// rate will increase
func (bf *Filter) Merge(oth *Filter) error {
	if bf.Size != oth.Size {
		return fmt.Errorf("incompatible filters size for merge : %d != %d",
			bf.Size, oth.Size)
	}
	if bf.k != oth.k {
		return fmt.Errorf("hashes functions must be the same to perform merge")
	}
	for i := uint64(0); i < bf.Size; i++ {
		bf.arr[i] |= oth.arr[i]
	}
	return nil
}

// Feed : Add an entry in the bloom filter
func (bf *Filter) Feed(s string) *Filter {
	hashs := bf.hasher([]byte(s))
	sectionSize := ((bf.Size * 8) / bf.k)
	for hid := uint64(0); hid < bf.k; hid++ {
		start := hid * sectionSize
		bf.setBit(start + (hashs[hid] % (sectionSize)))
	}
	bf.inserted++
	return bf
}

// FillRatio : Count each set bit into the Filter to compute the fillRatio
func (bf *Filter) FillRatio() float64 {
	return float64(popcntSliceGo(bf.arr)) / float64(bf.Size*8)
}

// EstimateFillRatio : Optimization of the fillRatio function, estimate instead of counting bits
func (bf *Filter) EstimateFillRatio() float64 {
	return 1.0 - math.Pow(
		(1.0-(1.0/((float64(bf.Size)*8.0)/float64(bf.k)))),
		float64(bf.inserted))
}
