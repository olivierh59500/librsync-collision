package main

import (
	"sort"
)

type Compressor struct {
	FreqCounts []FrequencyCount
	TScale     uint32
}

type FrequencyCount struct {
	Count    uint32
	Val      byte
	CodeBits uint32
	CodeVal  uint32
}

func NewCompressor(freqCounts []uint32) (cs Compressor) {
	cs.FreqCounts = make([]FrequencyCount, len(freqCounts))

	for i, count := range freqCounts {
		cs.FreqCounts[i] = FrequencyCount{Count: count, Val: byte(i)}
	}
	sort.Sort(sort.Reverse(ByFrequency(cs.FreqCounts)))
	return
}

func (self *Compressor) AssignCodes(tScale uint32) {
	var i uint32 = 0
	self.TScale = tScale
	for numBits := tScale; ; numBits += tScale {
		for val := uint32(0); val < (1 << numBits); val++ {
			if i >= uint32(len(self.FreqCounts)) {
				return
			}
			self.FreqCounts[i].CodeBits = numBits
			self.FreqCounts[i].CodeVal = val
			i++
		}
	}
}

func (self Compressor) EstimateSize() uint64 {
	var ret uint64
	for _, count := range self.FreqCounts {
		codeSize := uint64(count.CodeBits) * uint64(count.Count)
		ret += codeSize + (codeSize / uint64(self.TScale))
	}
	return (ret + 7) / 8
}

type ByFrequency []FrequencyCount

func (a ByFrequency) Len() int           { return len(a) }
func (a ByFrequency) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFrequency) Less(i, j int) bool { return a[i].Count < a[j].Count }
