package main

import (
	"fmt"
	"sync"
)

const f = 1
const N = 3*f + 1

type Message struct {
	Type    string
	Content string
}

type Party struct {
	id            int
	sentReady     bool
	mu            sync.Mutex
	receivedEcho  map[string]int
	receivedReady map[string]int
	outputChannel chan string
	isBad         bool
	bracha        *Bracha
	sendChannel   chan Message
}

type Bracha struct {
	parties []*Party
}

func NewBracha() *Bracha {
	bracha := &Bracha{}
	for i := 0; i < N; i++ {
		node := &Party{
			id:            i,
			receivedEcho:  make(map[string]int),
			receivedReady: make(map[string]int),
			bracha:        bracha,
			outputChannel: make(chan string),
			sendChannel:   make(chan Message),
		}
		bracha.parties = append(bracha.parties, node)
		go node.StartListener()
		go node.StartSender()
	}
	return bracha
}

func (p *Party) ReceiveMessage(m Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch m.Type {
	case "echo":
		p.receivedEcho[m.Content]++
		if p.receivedEcho[m.Content] >= 2*f+1 && !p.sentReady {
			p.sentReady = true
			p.Send(Message{Type: "ready", Content: m.Content})
		}

	case "ready":
		p.receivedReady[m.Content]++
		if p.receivedReady[m.Content] >= 2*f+1 {
			p.outputChannel <- m.Content
		}

	}
}

func (p *Party) Send(m Message) {
	p.sendChannel <- m
}

func (p *Party) StartSender() {
	for {
		select {
		case <-p.sendChannel:
			if p.isBad {
				p.Broadcast(Message{Type: "echo", Content: "Hello" + fmt.Sprint(666)})
			} else {
				p.Broadcast(Message{Type: "echo", Content: "Hello"})
			}
		}
	}
}

func (p *Party) StartListener() {
	for {
		select {
		case m := <-p.sendChannel:
			p.ReceiveMessage(m)
		}
	}
}

func (p *Party) Broadcast(m Message) {
	for i := 0; i < N; i++ {
		p.bracha.parties[i].Send(m)
	}
}

func main() {
	brachaInstance := NewBracha()

	brachaInstance.parties[2].isBad = true

	brachaInstance.parties[0].Send(Message{Type: "propose", Content: "Hello"})

	for _, party := range brachaInstance.parties {
		select {
		case output := <-party.outputChannel:
			if party.isBad {
				fmt.Println("Bad Party", party.id, "Output:", output)
			} else {
				fmt.Println("Good Party", party.id, "Output:", output)
			}
		default:
		}
	}
}
