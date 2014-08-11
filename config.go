package main

const (
	HASH_TRUNC     = 7         // 64-bit
	SPACE_WF       = (1 << 28) // ~ 80G of RAM
	TIME_WF        = (1 << 29) // SPACE_WF * TIME_WF should be >= (1<<64)
	STORE_PROCS    = 8
	GENERATE_PROCS = 8
	VERIFY_PROCS   = 8
	NUM_BUCKETS    = (SPACE_WF / STORE_PROCS * 4 / 3)
	NUM_HASHES     = 4
	CUCKOO_TTL     = 512
)
