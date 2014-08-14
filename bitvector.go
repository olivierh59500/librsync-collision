package main

type BitVector struct {
	vec []uint32
}

func (self BitVector) Get(start uint32, end uint32) uint32 {
	panic("TODO")
}

type SelectVector struct {
	vec    []uint32
	counts []uint8
}

func (self *SelectVector) SetBit(n uint32) {
	wordNum := n / 32
	bitNum := n % 32
	bitMask := uint32(1 << bitNum)

	if wordNum >= uint32(len(self.vec)) {
		oldVec := self.vec
		self.vec = make([]uint32, wordNum+1)
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

func (self *SelectVector) Rank(n uint32) uint32 {
	wordNum := n / 32
	bitNum := n % 32
	var i uint32
	var count uint32

	for i = 0; i < wordNum; i++ {
		count += uint32(self.counts[i])
	}

	word := self.vec[wordNum]
	for i = 0; i < bitNum; i++ {
		count += (word >> i) & 1
	}
	return count
}

func (self *SelectVector) Select(n uint32) uint32 {
	var count uint32
	var wordNum uint32
	for {
		if self.vec[wordNum]+count >= n {
			break
		}
		count += self.vec[wordNum]
		wordNum++
	}
	wordNum++

	lastWord := self.vec[wordNum]
	var i uint32
	for i = 0; count < n; i++ {
		count += (lastWord >> i) & 1
	}
	return i + (wordNum * 32)
}
