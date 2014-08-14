package main

import (
	"fmt"
	"github.com/dchest/siphash"
	"math/rand"
)

type HashTable struct {
	hashfunc HashFunc
	table    []byte
}

type HashFunc struct {
	K0 uint64
	K1 uint64
	M  uint64
}

func (self *HashFunc) Sum(data []byte) uint64 {
	return siphash.Hash(self.K0, self.K1, data) % self.M
}

func NewHashTable(num_buckets uint64) (table HashTable) {
	var r1 uint64 = (uint64(rand.Uint32()) << 32) | uint64(rand.Uint32())
	var r2 uint64 = (uint64(rand.Uint32()) << 32) | uint64(rand.Uint32())
	table.hashfunc = HashFunc{r1, r2, num_buckets}
	table.table = make([]byte, num_buckets*SEED_BYTES)
	return table
}

func (self *HashTable) Insert(val *DigestSeed) {
	offset := self.hashfunc.Sum(val.Digest) * SEED_BYTES
	copy(self.table[offset:], encode_seed(val.Seed))
}

func (self *HashTable) Get(key []byte) uint64 {
	h := self.hashfunc.Sum(key)
	start := h * SEED_BYTES
	end := start + SEED_BYTES
	return decode_seed(self.table[start:end])
}

func (self *HashTable) GetCompressedSize() (uint64, uint32) {
	freqCounts := make([]uint32, 256)
	for _, x := range self.table {
		freqCounts[x]++
	}
	cs := NewCompressor(freqCounts)

	minSize := uint64(0xffffffffffffffff)
	bestT := uint32(0)
	for i := uint32(1); i < 8; i *= 2 {
		cs.AssignCodes(i)
		sizeEstimate := cs.EstimateSize()
		if sizeEstimate < minSize {
			minSize, bestT = sizeEstimate, i
		}
		send_status(fmt.Sprintf("Size with tscale %v: %v", sizeEstimate, i))
	}
	cs.AssignCodes(bestT)

	return minSize, bestT
}

func (self *HashTable) GetUncompressedSize() uint64 {
	return uint64(len(self.table))
}

func make_hashtable(inchan <-chan *DigestSeed) HashTable {
	htable := NewHashTable(NUM_BUCKETS)
	var count uint64

	for data := range inchan {
		htable.Insert(data)
		if count++; count%REPORT_INTERVAL == 0 {
			send_status(fmt.Sprintf("Collected %d hashes", count))
		}
	}

	compressedSize, bestT := htable.GetCompressedSize()
	send_status(fmt.Sprintf("Approx compressed/uncompressed size: %v / %v [tscale=%v]", compressedSize, htable.GetUncompressedSize(), bestT))

	return htable
}

func bucket_finder(htable HashTable, inchan <-chan *DigestSeed, verify_chan chan<- Candidate) {
	for data := range inchan {
		seed := htable.Get(data.Digest)
		verify_chan <- Candidate{seed, data}
	}
}
