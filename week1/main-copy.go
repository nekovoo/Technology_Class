package main

import (
	"fmt"
)

type GoConsumer struct {
	msgs chan int
}

func GoNewConsumer(msgs chan int) *GoConsumer {
	return &GoConsumer{msgs: msgs}
}

func (c *GoConsumer) Goconsume() {
	fmt.Println("consume:Started")
	for {
		msgs := <-c.msgs
		fmt.Println("consume: Received:", msgs)
	}
}

type GoProducer struct {
	msgs chan int
	done *chan bool
}

func GoNewProducer(msgs chan int, done *chan bool) *GoProducer {
	return &GoProducer{msgs: msgs, done: done}
}

func (p *GoProducer) Goproduce(max int) {
	fmt.Println("produce: Started")
	for i := 0; i < max; i++ {
		fmt.Println("produce : Start sending ", i)
		p.msgs <- i
	}
	*p.done <- true
	fmt.Println("produce: Done")
}

func main() {
	var msgs = make(chan int)
	var done = make(chan bool)

	go GoNewProducer(msgs, &done).Goproduce(10)
	go GoNewConsumer(msgs).Goconsume()
	<-done
}
