package main // golang is so friggen annoying with it's packages

import (
	"fmt"
	"time" // Wish I could do this IRL
)

var (
	StatusChan chan StatusMsg
)

func send_status(msg string) {
	StatusChan <- StatusMsg{time.Now(), msg}
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

	result_chan := make(chan Result)
	verify_chan := make(chan Candidate)
	store_chans := make([]chan DigestSeed, STORE_PROCS)
	for i = 0; i < STORE_PROCS; i++ {
		store_ch := make(chan DigestSeed, 256)
		store_chans[i] = store_ch
		go start_storage_proc(store_ch, verify_chan)
	}

	for i = 0; i < GENERATE_PROCS; i++ {
		go generate_hashes(prefix1, prefix2, uint64(i), WORKFACTOR, GENERATE_PROCS, store_chans)
	}

	for i = 0; i < VERIFY_PROCS; i++ {
		go verify_collisions(prefix1, prefix2, verify_chan, result_chan)
	}

	result := <-result_chan
	send_status("Got result, exiting")
	close(StatusChan)
	<-status_finished

	return &result, true
}
