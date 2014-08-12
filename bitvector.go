package main

type BitVector struct {
	vec    []uint64
	counts []uint8
}

func (self *BitVector) SetBit(n uint64) {
	wordNum := n / 64
	bitNum := n % 64
	bitMask := uint64(1 << bitNum)

	if wordNum >= uint64(len(self.vec)) {
		oldVec := self.vec
		self.vec = make([]uint64, wordNum+1)
		copy(self.vec, oldVec)
		oldCounts := self.counts
		self.counts = make([]uint8, wordNum+1)
		copy(self.counts, oldCounts)
	}
	if self.vec[wordNum]&bitMask == 0 {
		self.counts[wordNum] += 1
	}
	self.vec[wordNum] |= bitMask
}

func (self *BitVector) Rank(n uint64) uint64 {
	wordNum := n / 64
	bitNum := n % 64
	var i uint64
	var count uint64

	for i = 0; i < wordNum; i++ {
		count += uint64(self.counts[i])
	}

	word := self.vec[wordNum]
	for i = 0; i < bitNum; i++ {
		count += (word >> i) & 1
	}
	return count
}

func (self *BitVector) Select(n uint64) uint64 {
	var count uint64
	var wordNum uint64
	for {
		if self.vec[wordNum]+count >= n {
			break
		}
		count += self.vec[wordNum]
		wordNum++
	}
	wordNum++

	lastWord := self.vec[wordNum]
	var i uint64
	for i = 0; count < n; i++ {
		count += (lastWord >> i) & 1
	}
	return i + (wordNum * 64)
}
