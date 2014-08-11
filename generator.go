package main

import (
	"github.com/therealmik/md4"
	"sync"
)

type PreparedHash struct {
	CtxTemplate *md4.Digest
}

func PrepareHash(prefix []byte) PreparedHash {
	ret := PreparedHash{md4.New()}
	ret.CtxTemplate.Write(prefix)
	return ret
}

func (self *PreparedHash) Hash(data []byte) []byte {
	ctx := self.CtxTemplate.Copy()
	ctx.Write(data)
	return ctx.Sum(nil)
}

func generate_hashes(prefix []byte, start uint64, stop uint64, step uint64, out_chans []chan *hash_and_input, wg *sync.WaitGroup) {
	h := PrepareHash(prefix)

	for i := start; i < stop; i += step {
		suffix := rollsum_expand(i)
		digest := truncate_hash(h.Hash(suffix))
		msg := &hash_and_input{Digest: digest, Seed: i}
		out_chans[int(digest[0])%STORE_PROCS] <- msg
	}
	if wg != nil {
		wg.Done()
	}
}
