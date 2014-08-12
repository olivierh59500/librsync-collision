package main

import (
	"encoding/gob"
	"fmt"
	"github.com/dchest/siphash"
	"io"
	"io/ioutil"
	"math/rand"
)

type HashTable struct {
	hashes HashFunc
	table  []DigestSeedHash
}

type HashFunc struct {
	K0 uint64
	K1 uint64
	M  uint64
}

type DigestSeedHash struct {
	Digest []byte
	Seed   uint64
	H      int
}

type CompressedTable struct {
	hashes HashFunc
	table  []uint32
}

func (self *HashFunc) Sum(data []byte) []uint64 {
	h := siphash.Hash(self.K0, self.K1, data)
	ret := make([]uint64, 3)
	ret[0] = (h >> 32) % self.M
	ret[1] = h % self.M
	ret[2] = ((h >> 32) + h) % self.M
	return ret
}

func NewHashTable(num_buckets uint64) (table HashTable) {
	var r1 uint64 = (uint64(rand.Uint32()) << 32) | uint64(rand.Uint32())
	var r2 uint64 = (uint64(rand.Uint32()) << 32) | uint64(rand.Uint32())
	table.hashes = HashFunc{r1, r2, num_buckets}
	table.table = make([]DigestSeedHash, num_buckets)
	return table
}

func (self *HashTable) displace(val *DigestSeedHash, ttl int) bool {
	if ttl <= 0 {
		return false
	}

	ok := true
	hs := self.hashes.Sum(val.Digest)
	h := hs[val.H%len(hs)]

	displaced := self.table[h]
	self.table[h] = *val
	if displaced.Digest != nil {
		displaced.H++
		ok = self.displace(&displaced, ttl-1)
	}

	return ok
}

func (self *HashTable) Insert(val *DigestSeedHash) bool {
	return self.displace(val, CUCKOO_TTL)
}

func (self *HashTable) Compress() CompressedTable {
	ret := CompressedTable{self.hashes, make([]uint32, len(self.table))}
	for i := 0; i < len(self.table); i++ {
		ret.table[i] = uint32(self.table[i].Seed)
	}
	return ret
}

func (self *CompressedTable) Get(key []byte) []uint32 {
	ret := make([]uint32, 3)
	for i, h := range self.hashes.Sum(key) {
		ret[i] = self.table[h]
	}
	return ret
}

func make_hashtable(inchan <-chan *DigestSeed) CompressedTable {
	tempfile, err := ioutil.TempFile(".", "hashstorage")
	if err != nil {
		panic(err)
	}

	var count uint64

	genc := gob.NewEncoder(tempfile)
	for data := range inchan {
		genc.Encode(*data)
		if count++; count&0xffffff == 0xffffff {
			send_status(fmt.Sprintf("Collected %d hashes", count))
		}
	}

	<-TableBuildSem

	if _, err := tempfile.Seek(0, 0); err != nil {
		panic("Unable to seek to the start of the tempfile")
	}

	htable := NewHashTable(NUM_BUCKETS)
	gdec := gob.NewDecoder(tempfile)
	count = 0
	var collision_count uint64
	for {
		data := DigestSeedHash{}
		if err := gdec.Decode(&data); err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		if ok := htable.Insert(&data); !ok {
			collision_count += 1
		}
		if count++; count&0xffffff == 0xffffff {
			send_status(fmt.Sprintf("Inserted %d hashes", count))
		}
	}
	send_status(fmt.Sprintf("Finished collecting hashes, %d received, %d collisions", count, collision_count))
	return htable.Compress()
}

func bucket_finder(htable CompressedTable, inchan <-chan *DigestSeed, verify_chan chan<- Candidate) {
	TableBuildSem <- struct{}{}
	for data := range inchan {
		for _, seed := range htable.Get(data.Digest) {
			verify_chan <- Candidate{uint64(seed), data}
		}
	}
}
