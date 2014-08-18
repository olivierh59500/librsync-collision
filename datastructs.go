package main

import "time"

type Result struct {
	Plaintext1 []byte
	Plaintext2 []byte
}

type StatusMsg struct {
	Timestamp time.Time
	Message   string
}

type DigestSeed struct {
	Digest [HASH_TRUNC]byte
	Seed   uint64
}

type Candidate struct {
	Seed1 uint64
	Seed2 uint64
}
