package main

import (
	"bytes"
	"fmt"
)

func verify_collisions(prefix []byte, verify_chan <-chan Candidate, result_chan chan<- Result) {
	h := PrepareHash(prefix)

	var count uint64
	for candidate := range verify_chan {
		suffix := rollsum_expand(candidate.Seed)
		digest := truncate_hash(h.Hash(suffix))
		if bytes.Equal(digest, candidate.Hash.Digest) {
			result_chan <- Result{candidate.Seed, candidate.Hash.Seed}
		}
		if count++; count%REPORT_INTERVAL == 0 {
			send_status(fmt.Sprintf("Tested %d hashes", count))
		}
	}
}
