package main

import "fmt"

func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	select { // ignores this because of default
	case c <- sum:
		{
			_ = 5 * 6
		}
	default:
		{
			_ = 4 * 7
		}
	}

	c <- 8
	<-c
	// send sum to c
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int) // Finds this one because no buffer size
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)

}
