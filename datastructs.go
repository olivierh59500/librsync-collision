package main

import "time"

type Result struct {
	Seed1 uint64
	Seed2 uint64
}

type StatusMsg struct {
	Timestamp time.Time
	Message   string
}

type DigestSeed struct {
	Digest []byte
	Seed   uint64
}

type Candidate struct {
	Seed uint64
	Hash *DigestSeed
}