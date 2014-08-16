package main

import (
	"github.com/therealmik/md4"
)

type PreparedHash struct {
	CtxTemplate *md4.Digest
}

func PrepareHash(prefix []byte) PreparedHash {
	ret := PreparedHash{md4.New()}
	ret.CtxTemplate.Write(prefix)
	return ret
}

func (self *PreparedHash) Hash(data [HASH_TRUNC]byte) [HASH_TRUNC]byte {
	ctx := self.CtxTemplate.Copy()
	ctx.Write(rollsum_encode(data))
	return truncate_hash(ctx.Sum(nil))
}

func is_distinguished(digest [HASH_TRUNC]byte) bool {
	for i := 0; i < LEADING_ZEROS; i++ {
		if digest[i] != 0 {
			return false
		}
	}
	return true
}

type Digester struct {
	H1       PreparedHash
	H2       PreparedHash
	FlagByte int
	FlagMask byte
}

func (self *Digester) Hash(data [HASH_TRUNC]byte) [HASH_TRUNC]byte {
	if data[self.FlagByte]&self.FlagMask == 0 {
		return self.H1.Hash(data)
	} else {
		return self.H2.Hash(data)
	}
}

func (self *Digester) WhichPrefix(data [HASH_TRUNC]byte) int {
	if data[self.FlagByte]&self.FlagMask == 0 {
		return 0
	} else {
		return 1
	}
}

func generate_hashes(prefix1 []byte, prefix2 []byte, start uint64, stop uint64, step uint64, out_chans []chan DigestSeed) {
	h := Digester{PrepareHash(prefix1), PrepareHash(prefix2), LEADING_ZEROS, 128}

	for i := start; i < stop; i += step {
		seed := encode_seed(i)
		var digest [HASH_TRUNC]byte = h.Hash(seed)
		for !is_distinguished(digest) {
			digest = h.Hash(digest)
		}
		msg := DigestSeed{Digest: digest, Seed: i}
		out_chans[int(digest[LEADING_ZEROS])%STORE_PROCS] <- msg
	}
}

func recreate_collision(prefix1 []byte, prefix2 []byte, candidate Candidate) ([HASH_TRUNC]byte, [HASH_TRUNC]byte, bool) {
	htable := make(map[[HASH_TRUNC]byte][HASH_TRUNC]byte)

	h := Digester{PrepareHash(prefix1), PrepareHash(prefix2), LEADING_ZEROS, 128}

	prev := encode_seed(candidate.Seed)
	for {
		digest := h.Hash(prev)
		htable[digest], prev = prev, digest
		if is_distinguished(digest) {
			break
		}
	}

	prev = encode_seed(candidate.Hash.Seed)
	for {
		digest := h.Hash(prev)
		collision, ok := htable[digest]
		if ok {
			p1 := h.WhichPrefix(collision)
			p2 := h.WhichPrefix(prev)
			if p1 == p2 {
				send_status("Collision with same prefix")
				break
			} else {
				send_status("Collision with different prefix")
				if p1 == 0 {
					return collision, prev, true
				} else {
					return prev, collision, true
				}
			}
		}
		if is_distinguished(digest) {
			break
		}
		prev = digest
	}

	send_status("Unable to recreate collision")
	return [HASH_TRUNC]byte{}, [HASH_TRUNC]byte{}, false
}
