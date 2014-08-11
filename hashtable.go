package main

import (
	"fmt"
	"github.com/dchest/siphash"
	"math/rand"
)

type HashTable struct {
	hashes []HashFunc
	table  []hash_and_input
}

type HashFunc struct {
	K0 uint64
	K1 uint64
	M  uint64
}

type CompressedTable struct {
	hashes []HashFunc
	table  []uint32
}

func (self *HashFunc) Sum(data []byte) uint64 {
	return siphash.Hash(self.K0, self.K1, data) % self.M
}

func NewHashTable(num_hashes uint64, num_buckets uint64) (table HashTable) {
	var i uint64
	table.hashes = make([]HashFunc, num_hashes)
	for i = 0; i < num_hashes; i++ {
		var r uint64 = (uint64(rand.Uint32()) << 32) | uint64(rand.Uint32())
		table.hashes[i] = HashFunc{r, uint64(i), num_buckets}
	}
	table.table = make([]hash_and_input, num_buckets)
	return table
}

func (self *HashTable) displace(val *hash_and_input, hashid int, ttl int) bool {
	if ttl <= 0 {
		return false
	}

	ok := true
	h := self.hashes[hashid].Sum(val.Digest)

	displaced := self.table[h]
	if displaced.Digest != nil {
		ok = self.displace(&displaced, (hashid+1)%len(self.hashes), ttl-1)
	}
	self.table[h] = *val

	return ok
}

func (self *HashTable) Insert(val *hash_and_input) bool {
	return self.displace(val, 0, CUCKOO_TTL)
}

func (self *HashTable) Compress() CompressedTable {
	ret := CompressedTable{self.hashes, make([]uint32, len(self.table))}
	for i := 0; i < len(self.table); i++ {
		ret.table[i] = uint32(self.table[i].Seed)
	}
	return ret
}

func (self *CompressedTable) Get(key []byte) []uint32 {
	ret := make([]uint32, len(self.hashes))
	for i, hash := range self.hashes {
		h := hash.Sum(key)
		ret[i] = self.table[h]
	}
	return ret
}

func make_hashtable(inchan <-chan *hash_and_input) CompressedTable {
	htable := NewHashTable(NUM_HASHES, NUM_BUCKETS)

	var count uint64
	var collision_count uint64

	for data := range inchan {
		if ok := htable.Insert(data); !ok {
			collision_count += 1
		}
		if count++; count&0xffffff == 0xffffff {
			send_status(fmt.Sprintf("Collected %d hashes", count))
		}
	}
	send_status(fmt.Sprintf("Finished collecting hashes, %d received, %d collisions", count, collision_count))
	return htable.Compress()
}

func bucket_finder(htable CompressedTable, inchan <-chan *hash_and_input, verify_chan chan<- Candidate) {
	for data := range inchan {
		for _, seed := range htable.Get(data.Digest) {
			verify_chan <- Candidate{uint64(seed), data}
		}
	}
}
