package main

import (
	"github.com/therealmik/md4"
	"sync"
)

func generate_hashes(prefix []byte, start uint64, stop uint64, step uint64, out_chans []chan *hash_and_input, wg *sync.WaitGroup) {
	ctxtmpl := md4.New()
	ctxtmpl.Write(prefix)

	for i := start; i < stop; i += step {
		ctx := ctxtmpl.Copy()
		ctx.Write(rollsum_expand(i))
		digest := truncate_hash(ctx.Sum(nil))
		msg := &hash_and_input{Digest: digest, Seed: i}
		out_chans[int(digest[0])%STORE_PROCS] <- msg
	}
	if wg != nil {
		wg.Done()
	}
}
