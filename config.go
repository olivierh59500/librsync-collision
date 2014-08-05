package main

const (
	HASH_TRUNC  = 8                     // 64-bit
	SPACE_WF    = (1 << 32) - (1 << 30) // ~ 80G of RAM
	TIME_WF     = (1 << 34)             // SPACE_WF * TIME_WF should be >= (1<<64)
	STORE_PROCS = 4
)
