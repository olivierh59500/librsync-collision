package main

import (
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
	Seed   []byte
	H      int
}

type CompressedTable struct {
	hashes HashFunc
	table  []byte
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
	ret := CompressedTable{self.hashes, make([]byte, len(self.table)*SEED_BYTES)}
	for i := 0; i < len(self.table); i++ {
		if self.table[i].Seed != nil {
			copy(ret.table[i*SEED_BYTES:], self.table[i].Seed)
		}
	}
	return ret
}

func (self *CompressedTable) Get(key []byte) []uint64 {
	ret := make([]uint64, 3)
	for i, h := range self.hashes.Sum(key) {
		start := h * SEED_BYTES
		end := start + SEED_BYTES
		ret[i] = decode_seed(self.table[start:end])
	}
	return ret
}

func read_from_tempfile(tempfile io.Reader, ch chan<- DigestSeedHash) {
	defer close(ch)

	for {
		buf := make([]byte, HASH_TRUNC+SEED_BYTES)
		var n int = 0
		var err error
		for n < HASH_TRUNC+SEED_BYTES && err == nil {
			var m int
			m, err = tempfile.Read(buf[n:])
			n += m
		}
		if n == HASH_TRUNC+SEED_BYTES {
			ch <- DigestSeedHash{Digest: buf[:HASH_TRUNC], Seed: buf[HASH_TRUNC:]}
		}
		if err != nil {
			return
		}
	}
}

func make_hashtable(inchan <-chan *DigestSeed) CompressedTable {
	tempfile, err := ioutil.TempFile(".", "hashstorage")
	if err != nil {
		panic(err)
	}

	var count uint64

	freqCounts := make([]uint32, 256)

	for data := range inchan {
		seed := encode_seed(data.Seed)
		for _, x := range(seed) {
			freqCounts[x]++
		}

		buf := make([]byte, HASH_TRUNC+SEED_BYTES)
		copy(buf, data.Digest)
		copy(buf[HASH_TRUNC:], seed)
		if _, err := tempfile.Write(buf); err != nil {
			panic("Unable to write to tempfile")
		}

		if count++; count&0xfffff == 0xfffff {
			send_status(fmt.Sprintf("Collected %d hashes", count))
		}
	}

	<-TableBuildSem
	send_status(fmt.Sprintf("Frequency Counts: %v", freqCounts))

	if _, err := tempfile.Seek(0, 0); err != nil {
		panic("Unable to seek to the start of the tempfile")
	}

	htable := NewHashTable(NUM_BUCKETS)
	count = 0
	var collision_count uint64
	ch := make(chan DigestSeedHash, (1 << 8))
	go read_from_tempfile(tempfile, ch)

	for data := range ch {
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
			verify_chan <- Candidate{seed, data}
		}
	}
}
