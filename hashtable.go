package main

import (
	"fmt"
)

type HashTable map[[HASH_TRUNC]byte]uint64

func start_storage_proc(inchan <-chan DigestSeed, verify_chan chan<- Candidate) {
	htable := make(HashTable, NUM_BUCKETS)
	var count uint64

	for data := range inchan {
		collision, ok := htable[data.Digest]
		if ok {
			verify_chan <- Candidate{collision, data.Seed}
		} else {
			htable[data.Digest] = data.Seed
		}
		if count++; count%REPORT_INTERVAL == 0 {
			send_status(fmt.Sprintf("Collected %d hashes", count))
		}
	}
}
