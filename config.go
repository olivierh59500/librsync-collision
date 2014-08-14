package main

const (
	HASH_TRUNC     = 6         // 64-bit
	SPACE_WF       = (1 << 24) // ~ 80G of RAM
	TIME_WF        = (1 << 34) // SPACE_WF * TIME_WF should be >= (1<<64)
	STORE_PROCS    = 8
	GENERATE_PROCS = 8
	VERIFY_PROCS   = 8
	NUM_BUCKETS    = (SPACE_WF / STORE_PROCS)
	SEED_BYTES     = 3
)
