package main // golang is so friggen annoying with it's packages

import (
	"fmt"
	"sync"
	"time" // Wish I could do this IRL
)

var (
	StatusChan    chan StatusMsg
	TableBuildSem chan struct{}
)

func send_status(msg string) {
	StatusChan <- StatusMsg{time.Now(), msg}
}

func start_storage_proc(store_chan <-chan *DigestSeed, test_chan <-chan *DigestSeed, verify_chan chan<- Candidate) {
	table := make_hashtable(store_chan)
	bucket_finder(table, test_chan, verify_chan)
}

func status_printer(finished chan<- struct{}) {
	start_time := time.Now()
	for msg := range StatusChan {
		fmt.Println(msg.Timestamp.Sub(start_time), msg.Message)
	}
	finished <- struct{}{}
}

func Run(prefix1, prefix2 []byte) (*Result, bool) {
	var i int

	if rollsum(prefix1) != rollsum(prefix2) {
		fmt.Println("Rolling sums for prefixes don't match")
		return nil, false
	}

	// Print status/progress messages in a goroutine without blocking work
	StatusChan = make(chan StatusMsg, 16)
	status_finished := make(chan struct{})
	go status_printer(status_finished)

	TableBuildSem = make(chan struct{}, SIMUL_TABLE_BUILDS)
	for i = 0; i < SIMUL_TABLE_BUILDS; i++ {
		TableBuildSem <- struct{}{}
	}

	result_chan := make(chan Result)
	verify_chan := make(chan Candidate, (1 << 16))
	store_chans := make([]chan *DigestSeed, STORE_PROCS)
	test_chans := make([]chan *DigestSeed, STORE_PROCS)
	for i = 0; i < STORE_PROCS; i++ {
		store_ch := make(chan *DigestSeed, 1<<16)
		test_ch := make(chan *DigestSeed, 1<<16)
		store_chans[i] = store_ch
		test_chans[i] = test_ch
		go start_storage_proc(store_ch, test_ch, verify_chan)
	}

	var store_wg sync.WaitGroup
	store_wg.Add(GENERATE_PROCS)
	for i = 0; i < GENERATE_PROCS; i++ {
		go generate_hashes(prefix1, uint64(i), SPACE_WF, GENERATE_PROCS, store_chans, &store_wg)
		go generate_hashes(prefix2, uint64(i), TIME_WF, GENERATE_PROCS, test_chans, nil)
	}
	store_wg.Wait()
	for i = 0; i < STORE_PROCS; i++ {
		close(store_chans[i])
	}

	for i = 0; i < VERIFY_PROCS; i++ {
		go verify_collisions(prefix1, verify_chan, result_chan)
	}

	result := <-result_chan
	send_status(fmt.Sprintf("Got result: %d %d", result.Seed1, result.Seed2))
	close(StatusChan)
	<-status_finished

	return &result, true
}
