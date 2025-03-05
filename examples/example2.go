package main

import (
	"fmt"
)

func main2() {
	ch := make(chan int)
	ch2 := make(chan int)
	var chInterface interface{} = ch2

	select {

	// Valid
	case <-ch:
		fmt.Println("Received from ch")
	// Valid
	case chInterface.(chan int) <- 3: // Type assertion
		fmt.Println("Sent to ch2 via interface")
	// Valid
	case <-func() chan int { return ch }(): // Function literal returning a channel
		fmt.Println("Received from ch via function literal")
	// Valid
	default:
		fmt.Println("Default case")
	}
}
