// Package bloom package implement a simple and scalable Bloom Filter algorithm
package bloom

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
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

// EncodedFilter is the JSON filter structure
type EncodedFilter struct {
	Arr      []byte
	Size     uint64
	K        uint64
	Inserted uint64
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
	enc := &EncodedFilter{
		Size:     bf.Size,
		Arr:      bf.arr,
		K:        bf.k,
		Inserted: bf.inserted,
	}

	return json.Marshal(enc)
}

// ToFile : Export filter to a file
func (bf *Filter) ToFile(path string) error {
	json, err := bf.ToJSON()
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(json)
	return err
}

// FromFile : Import filter from a file
func FromFile(path string) (*Filter, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return FromJSON(bytes)
}

// FromJSON : Import a JSON serialized bloom filter
func FromJSON(raw []byte) (*Filter, error) {
	var dat map[string]interface{}

	if err := json.Unmarshal(raw, &dat); err != nil {
		return nil, err
	}
	bf := New(uint64(dat["Size"].(float64)), uint64(dat["K"].(float64)))
	bf.inserted = uint64(dat["Inserted"].(float64))
	n, err := base64.StdEncoding.DecodeString(dat["Arr"].(string))
	if err != nil {
		return nil, err
	}
	bf.arr = n
	return bf, nil
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
