package main

import (
	"fmt"
	"time"
)

func main4() {
	ch := make(chan int, 100)
	var chInterface interface{} = ch

	select {

	// Valid
	case chInterface.(chan int) <- 100: // Type assertion
		fmt.Println("Sent to ch2 via interface")
	// Valid <---- need to catch this
	case <-(time.After(500 * time.Millisecond)):
		fmt.Println("Timeout")
	}
}
