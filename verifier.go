package main

import (
	"bytes"
	"fmt"
	"github.com/therealmik/md4"
)

func verify_collisions(prefix []byte, verify_chan <-chan Candidate, result_chan chan<- Result) {
	ctxtmpl := md4.New()
	ctxtmpl.Write(prefix)

	var count uint64
	for candidate := range verify_chan {
		for _, input := range candidate.Bucket {
			ctx := ctxtmpl.Copy()
			ctx.Write(rollsum_expand(input))
			digest := truncate_hash(ctx.Sum(nil))
			if bytes.Equal(digest, candidate.Hash.Digest) {
				result_chan <- Result{input, candidate.Hash.Seed}
			}
		}
		if count++; count&0xfffff == 0xfffff {
			send_status(fmt.Sprintf("Tested %d hashes", count))
		}
	}
}
