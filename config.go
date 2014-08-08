package main

const (
	HASH_TRUNC       = 7         // 64-bit
	SPACE_WF         = (1 << 32) // ~ 80G of RAM
	TIME_WF          = (1 << 33) // SPACE_WF * TIME_WF should be >= (1<<64)
	STORE_PROCS      = 4
	GENERATE_PROCS   = 6
	VERIFY_PROCS     = 6
	ITEMS_PER_BUCKET = 2 // Average items per bucket
	NUM_BUCKETS      = (SPACE_WF / STORE_PROCS / ITEMS_PER_BUCKET)
)
