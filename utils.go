package main

import (
	"hash/fnv"
)

func truncate_hash(input []byte) []byte {
	return input[:HASH_TRUNC]
}

func rollsum_expand(n uint64) []byte {
	ret := make([]byte, 20)
	var i uint

	for i = 0; i < 5; i++ {
		b := byte((n >> (i * 8)) & 0xff)
		c := 255 - b
		ret[(i*4)+0] = b
		ret[(i*4)+1] = c
		ret[(i*4)+2] = c
		ret[(i*4)+3] = b
	}
	return ret
}

func rollsum(data []byte) uint32 {
	var A uint32
	var B uint32

	for i, x := range data {
		A += uint32(x)
		B += uint32(len(data)-i) * uint32(x)
	}
	return (A & 0xffff) | ((B & 0xffff) << 16)
}

func hash(key []byte) uint32 {
	ctx := fnv.New32a()
	ctx.Write(key)
	return ctx.Sum32() % NUM_BUCKETS
}
