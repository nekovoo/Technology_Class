package main

import (
	"fmt"
	"log"
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
	isLocked      bool
	content       string
}

type Bracha struct {
	parties []*Party
}

func NewBracha() *Bracha {
	bracha := &Bracha{}
	for i := 0; i < N; i++ {
		node := &Party{
			isBad:         false,
			id:            i,
			receivedEcho:  make(map[string]int),
			receivedReady: make(map[string]int),
			bracha:        bracha,
			outputChannel: make(chan string),
			isLocked:      false,
			content:       "",
		}
		bracha.parties = append(bracha.parties, node)
	}
	return bracha
}

func (p *Party) ReceiveMessage(m Message) {

	p.mu.Lock()
	p.isLocked = true
	var broadcastMessage *Message = nil
	switch m.Type {
	case "propose":
		if p.isBad {
			broadcastMessage = &Message{Type: "echo", Content: m.Content + fmt.Sprint(666)}
		} else {
			broadcastMessage = &Message{Type: "echo", Content: m.Content}
		}

	case "echo":

		p.receivedEcho[m.Content]++
		log.Printf("p%d receiveEcho[%s] = %d", p.id, m.Content, p.receivedEcho[m.Content])
		if p.receivedEcho[m.Content] >= 2*f+1 && !p.sentReady {
			p.sentReady = true
			broadcastMessage = &Message{Type: "ready", Content: m.Content}
		}

	case "ready":

		p.receivedReady[m.Content]++
		//log.Printf("p%d receiveReady[%s] = %d", p.id, m.Content, p.receivedReady[m.Content])
		if p.receivedReady[m.Content] >= 2*f+1 {
			//fmt.Printf("p%d is %s\n", p.id, m.Content)
			p.content = m.Content
		}

	}
	p.mu.Unlock()
	p.isLocked = false
	if broadcastMessage != nil {
		p.Broadcast(*broadcastMessage, p.bracha)
	}

}

func (p *Party) Broadcast(m Message, bracha *Bracha) {
	for i := 0; i < N; i++ {
		if bracha.parties[i].isLocked {
			bracha.parties[i].mu.Unlock()
		}
	}
	for i := 0; i < N; i++ {
		if p.isLocked {
			p.mu.Unlock()
		}
		if bracha.parties[i].isLocked {
			bracha.parties[i].mu.Unlock()
		}
		//log.Printf("p%d broadcast %s to p%d", p.id, m.Type, bracha.parties[i].id)
		bracha.parties[i].ReceiveMessage(m)
		//log.Printf("parties%d recive %s", bracha.parties[i].id, m.Type)
	}
}
func (p *Party) Broadcast_start(m Message, bracha *Bracha) {

}
func main() {
	brachaInstance := NewBracha()

	// Set the first Party as a malicious node
	brachaInstance.parties[2].isBad = true

	// Broadcast a proposal message from the first Party
	brachaInstance.parties[0].Broadcast(Message{Type: "propose", Content: "Hello"}, brachaInstance)
	for _, party := range brachaInstance.parties {
		fmt.Printf("p[%d] is %s\n", party.id, party.content)
	}
	/*for _, party := range brachaInstance.parties {
		select {
		case output := <-party.outputChannel:
			if party.isBad {
				fmt.Println("Bad Party", party.id, "Output:", output)
			} else {
				fmt.Println("Good Party", party.id, "Output:", output)
			}
		default:
		}
	}*/
}
