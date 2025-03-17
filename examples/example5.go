package main

import (
	"fmt"
	"time"
)

// Doesn't detect this currently
var channelAmount = 100

func main5() {
	ch := make(chan int, channelAmount)
	var chInterface interface{} = ch

	select {

	// Valid
	case chInterface.(chan int) <- 3: // Type assertion
		fmt.Println("Sent to ch2 via interface")
	// Valid <---- need to catch this
	case <-time.After(500 * time.Millisecond):
		fmt.Println("Timeout")
	}

	// These are all cases it doesn't catch :(
	d := 500 * time.Millisecond
	var timer interface{} = time.NewTimer(d).C
	select {
	case ch <- 0:
		// ...
	case <-timer.(<-chan time.Time): // Type assertion
		// ...
	}

	timer2 := time.After(d)
	select {
	case ch <- 1:
		// ...
	case <-timer2:
		// ...
	}

	select {
	case ch <- 2:
		// ...
	case <-getTimer(d):
		// ...
	}

	select {
	case ch <- 3:
		// ...
	case <-wrapAfter(d):
		// ...
	}

	timer3 := time.NewTimer(500)
	select {
	case ch <- 4:
		// ...
	case <-timer3.C:
		// ...
	}
}

func getTimer(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func wrapAfter(d time.Duration) <-chan time.Time {
	return time.After(d)
}
