package main // golang is so friggen annoying with it's packages

import (
	"fmt"
	"github.com/therealmik/md4"
	"sync"
	"time" // Wish I could do this IRL
)

type Result struct {
	Input1 uint64
	Input2 uint64
}

type StatusMsg struct {
	Timestamp time.Time
	Message   string
}

type truncated_hash [HASH_TRUNC]byte

type hash_and_input struct {
	Digest truncated_hash
	Input  uint64
}

func send_status(ch chan StatusMsg, msg string) {
	ch <- StatusMsg{time.Now(), msg}
}

func truncate_hash(input []byte) truncated_hash {
	var ret truncated_hash
	for i, x := range input[:HASH_TRUNC] {
		ret[i] = x
	}
	return ret
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

func make_hashtable(inchan chan *hash_and_input, status_chan chan StatusMsg) map[truncated_hash]uint32 {
	collision_map := make(map[truncated_hash]uint32, SPACE_WF/STORE_PROCS)
	var count uint64

	for data := range inchan {
		collision_map[data.Digest] = uint32(data.Input)
		if count++; count&0xffffff == 0xffffff {
			send_status(status_chan, fmt.Sprintf("Collected %d hashes", count))
		}
	}
	send_status(status_chan, fmt.Sprintf("Finished collecting hashes, %d entries", len(collision_map)))
	return collision_map
}

func test_for_collisions(collision_map map[truncated_hash]uint32, inchan chan *hash_and_input, statuschan chan StatusMsg, resultchan chan Result) {
	var count uint64

	for data := range inchan {
		CollInput, ok := collision_map[data.Digest]
		if ok {
			resultchan <- Result{uint64(CollInput), data.Input}
			break
		}
		count += 1
		if count&0xffffff == 0xffffff {
			send_status(statuschan, fmt.Sprintf("Tested %d hashes", count))
		}
	}
}

func generate_hashes(prefix []byte, start uint64, stop uint64, step uint64, out_chans []chan *hash_and_input, wg *sync.WaitGroup) {
	ctxtmpl := md4.New()
	ctxtmpl.Write(prefix)

	for i := start; i < stop; i += step {
		ctx := ctxtmpl.Copy()
		ctx.Write(rollsum_expand(i))
		digest := truncate_hash(ctx.Sum(nil))
		msg := &hash_and_input{Digest: digest, Input: i}
		out_chans[int(digest[0])%STORE_PROCS] <- msg
	}
	if wg != nil {
		wg.Done()
	}
}

func start_storage_proc(store_chan chan *hash_and_input, test_chan chan *hash_and_input, status_chan chan StatusMsg, result_chan chan Result) {
	table := make_hashtable(store_chan, status_chan)
	test_for_collisions(table, test_chan, status_chan, result_chan)
}

func status_printer(status_chan chan StatusMsg, finished chan struct{}) {
	start_time := time.Now()
	for msg := range status_chan {
		fmt.Println(msg.Timestamp.Sub(start_time), msg.Message)
	}
	finished <- struct{}{}
}

func Run(prefix1, prefix2 []byte, numThreads int) (*Result, bool) {
	var i int

	if rollsum(prefix1) != rollsum(prefix2) {
		fmt.Println("Rolling sums for prefixes don't match")
		return nil, false
	}

	// Print status/progress messages in a goroutine without blocking work
	status_chan := make(chan StatusMsg, (STORE_PROCS+numThreads)*2)
	status_finished := make(chan struct{})
	go status_printer(status_chan, status_finished)

	result_chan := make(chan Result)
	store_chans := make([]chan *hash_and_input, STORE_PROCS)
	test_chans := make([]chan *hash_and_input, STORE_PROCS)
	for i = 0; i < STORE_PROCS; i++ {
		store_ch := make(chan *hash_and_input, 1<<16)
		test_ch := make(chan *hash_and_input, 1<<16)
		store_chans[i] = store_ch
		test_chans[i] = test_ch
		go start_storage_proc(store_ch, test_ch, status_chan, result_chan)
	}

	var store_wg sync.WaitGroup
	store_wg.Add(numThreads)
	for i = 0; i < numThreads; i++ {
		go generate_hashes(prefix1, uint64(i), SPACE_WF, uint64(numThreads), store_chans, &store_wg)
		go generate_hashes(prefix2, uint64(i), TIME_WF, uint64(numThreads), test_chans, nil)
	}
	store_wg.Wait()
	for i = 0; i < STORE_PROCS; i++ {
		close(store_chans[i])
	}

	result := <-result_chan
	send_status(status_chan, fmt.Sprintf("Got result: %d %d", result.Input1, result.Input2))
	close(status_chan)
	<-status_finished

	return &result, true
}
