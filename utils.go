package bloom

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"math"
)

// Bug(makeHashes) Panic if required hashes total size is bigger than a sha512

// Panic if required hashes are larger than a sha512 digest
// Final hash function, compute k hashes of size characters from
// digest.
func makeHashes(digest []byte, size uint64, k uint64) []uint64 {
	if uint64(len(digest)) < size*k || size > 8 {
		panic("Digest is too small to address all the filter")
	}
	res := make([]uint64, k)
	for i := uint64(0); i < k; i++ {
		raw := digest[i*size : ((i + 1) * size)]
		res[i] = binary.BigEndian.Uint64(append(make([]byte, 8-len(raw)), raw...))
	}
	return res
}

// Return the function used in the filter for hashing
//
func hashingRoutine(size uint64, k uint64) func([]byte) []uint64 {
	hashSize := size * k

	return func(inp []byte) []uint64 {

		switch {
		case hashSize > 48:
			digest := (sha512.Sum512(inp))
			return makeHashes(digest[:], size, k)
		case hashSize > 32:
			digest := (sha512.Sum384(inp))
			return makeHashes(digest[:], size, k)
		case hashSize > 20:
			digest := (sha256.Sum256(inp))
			return makeHashes(digest[:], size, k)
		case hashSize > 16:
			digest := (sha1.Sum(inp))
			return makeHashes(digest[:], size, k)
		default:
			digest := (md5.Sum(inp))
			return makeHashes(digest[:], size, k)
		}
	}
}

// M bits, k hashers
//
func generateHasher(k uint64, M uint64) func([]byte) [](uint64) {
	minIdx := (M / k)
	minHashDigits := uint64(math.Ceil(math.Log(float64(minIdx)) / math.Log(16.0)))
	return hashingRoutine(minHashDigits, k)
}

// n : number to test, b : base
// func digitNum(n uint64, b uint64) uint64 {
// 	return uint64(math.Ceil(math.Log(float64(n)) / math.Log(float64(b))))
// }
//
// // Number of digits in an hexadecimal notation number
// func hexDigits(n uint64) uint64 {
// 	return digitNum(n, 16)
// }

// Compute k for a determined false positive rate
func hashCountForFP(fp float64) uint64 {
	return uint64(math.Ceil(math.Log2(1.0 / fp)))
}

// set nth bit to 1
func (bf *Filter) setBit(n uint64) *Filter {
	bf.arr[n/8] |= (1 << (n % 8))
	return bf
}

// n : bit index
func (bf *Filter) isSet(n uint64) bool {
	return (bf.arr[n/8] & (1 << (n % 8))) > 0
}

// Estimate gives an estimation of optimal configuration for
// A bloom filter with n elements for a false positive probability
// desired of fp %
// // https://en.wikipedia.org/wiki/Bloom_filter#Optimal_number_of_hash_functions
// func Estimate(fp float64) (k uint64, m uint64) {
// 	// m = uint64(math.Ceil(-(float64(n) * math.Log2(fp/100.0) / (math.Ln2 * math.Ln2))))
// 	// k = uint64(math.Ceil(float64(m/n) * math.Ln2))
// 	k = uint64(math.Ceil(math.Log2(1.0 / fp)))
//
// 	return
// }

// bit population count, taken from
// https://code.google.com/p/go/issues/detail?id=4988#c11
// credit: https://code.google.com/u/arnehormann/
func popcount(x uint64) (n uint64) {
	x -= (x >> 1) & 0x5555555555555555
	x = (x>>2)&0x3333333333333333 + x&0x3333333333333333
	x += x >> 4
	x &= 0x0f0f0f0f0f0f0f0f
	x *= 0x0101010101010101
	return x >> 56
}

func popcntSliceGo(s []byte) uint64 {
	cnt := uint64(0)
	for _, x := range s {
		cnt += popcount(uint64(x))
	}
	return cnt
}
