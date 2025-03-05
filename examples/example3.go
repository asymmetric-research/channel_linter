package main

import (
	"fmt"
)

func main3() {
	ch := make(chan int, 2)
	ch2 := make(chan int, 2)
	ch3 := make(chan int)
	chFunc := func() chan int { return ch3 }
	var chInterface interface{} = ch2
	channels := []chan int{ch, ch2}

	select {

	// Valid statement
	case <-ch:
		fmt.Println("Received from ch")

	// Valid statement
	case chInterface.(chan int) <- 3: // Type assertion
		fmt.Println("Sent to ch2 via interface")

	// Valid statement
	case channels[0] <- 4: // Index expression
		fmt.Println("Sent to ch via index expression")

	// Valid statement
	case <-func() chan int { return ch }(): // Function literal returning a channel
		fmt.Println("Received from ch via function literal")

	// Valid statement
	case chFunc() <- 5: // Function call returning a channel
		fmt.Println("Sent to ch3 via function call")

	default:
		fmt.Println("Default case")
	}
}
