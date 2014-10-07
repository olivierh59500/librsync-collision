package main

func verify_collisions(prefix1 []byte, prefix2 []byte, verify_chan <-chan Candidate, result_chan chan<- Result) {
	for candidate := range verify_chan {
		send_status("Verifying collision...")
		digest1, digest2, ok := recreate_collision(prefix1, prefix2, candidate)
		if ok {
			send_status("Got result")
			result_chan <- MakeResult(prefix1, prefix2, digest1, digest2)
		}
	}
}

func MakeResult(prefix1, prefix2 []byte, digest1 [HASH_TRUNC]byte, digest2 [HASH_TRUNC]byte) Result {
	result1 := append(prefix1, rollsum_encode(digest1)...)
	result2 := append(prefix2, rollsum_encode(digest2)...)
	return Result{result1, result2}
}
