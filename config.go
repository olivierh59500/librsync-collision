package main

const (
	HASH_TRUNC      = 8 // 64-bit
	WORKFACTOR      = (1 << 26)
	STORE_PROCS     = 2
	GENERATE_PROCS  = 8
	VERIFY_PROCS    = 1
	NUM_BUCKETS     = (WORKFACTOR / STORE_PROCS) // Keep the table dense, even if we lose some keys
	SEED_BYTES      = HASH_TRUNC
	REPORT_INTERVAL = 1024
	LEADING_ZEROS   = 2
)
