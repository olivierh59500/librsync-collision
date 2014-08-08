package main

import "fmt"

func make_hashtable(inchan <-chan *hash_and_input) []uint32 {
	collision_map := make([]uint32, NUM_BUCKETS)
	var count uint64

	for data := range inchan {
		bucket := hash(data.Digest)
		collision_map[bucket] = uint32(data.Seed)
		if count++; count&0xffffff == 0xffffff {
			send_status(fmt.Sprintf("Collected %d hashes", count))
		}
	}
	send_status(fmt.Sprintf("Finished collecting hashes, %d entries", len(collision_map)))
	return collision_map
}

func bucket_finder(collision_map []uint32, inchan <-chan *hash_and_input, verify_chan chan<- Candidate) {
	for data := range inchan {
		seed := uint64(collision_map[hash(data.Digest)])
		if seed != 0 {
			verify_chan <- Candidate{seed, data}
		}
	}
}
