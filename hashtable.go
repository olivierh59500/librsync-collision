package main

import "fmt"

func make_hashtable(inchan <-chan *hash_and_input) [][]uint64 {
	collision_map := make([][]uint64, NUM_BUCKETS)
	var count uint64

	for data := range inchan {
		bucket := hash(data.Digest)
		collision_map[bucket] = append(collision_map[bucket], data.Seed)
		if count++; count&0xffffff == 0xffffff {
			send_status(fmt.Sprintf("Collected %d hashes", count))
		}
	}
	send_status(fmt.Sprintf("Finished collecting hashes, %d entries", len(collision_map)))
	return collision_map
}

func bucket_finder(collision_map [][]uint64, inchan <-chan *hash_and_input, verify_chan chan<- Candidate) {
	for data := range inchan {
		bucket := collision_map[hash(data.Digest)]
		verify_chan <- Candidate{bucket, data}
	}
}
