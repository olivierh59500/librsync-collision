package main

const (
	HASH_TRUNC      = 8 // 64-bit
	WORKFACTOR      = (1 << 26)
	STORE_PROCS     = 2
	GENERATE_PROCS  = 8
	VERIFY_PROCS    = 1
	NUM_BUCKETS     = (1 << 17) // Don't waste too much memory, but big enough to hold our dataset easily
	SEED_BYTES      = HASH_TRUNC
	REPORT_INTERVAL = 1024
	LEADING_ZEROS   = 2
)
