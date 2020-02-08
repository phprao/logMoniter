package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string)
	go t2(ch1)
	go t1(ch1)

	time.Sleep(1 * time.Second)
}

func t1(ch chan string){
	for v := range ch {
		fmt.Println(v)
	}
}

func t2(ch chan string){
	ch <- "a"
	ch <- "b"
	ch <- "c\n"
	ch <- "d"
}