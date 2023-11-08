package aab

import (
	"Dory/internal/party"
	"Dory/pkg/protobuf"
	"sync"
)

//var mm map[uint32]chan bool
func isBad(id uint32, F uint32) bool {
	if id > 2*F {
		return true
	}
	return false
}
func int2bytes(i uint32) []byte {
	b1 := byte(i)
	b2 := byte(i >> 8)
	b3 := byte(i >> 16)
	b4 := byte(i >> 24)

	return []byte{b4, b3, b2, b1}
}

func RBCMain(p *party.HonestParty, inputChannel chan []byte, outputChannel chan []byte, f int, pid int) {
	var idx uint32
	idx = 0
	mm := make(map[uint32]chan bool)
	mm[0] = make(chan bool, 1)
	mm[0] <- true

	for {
		idx++
		id := int2bytes(idx)

		mm[idx] = make(chan bool, 1)
		data, _ := p.GetMessageWithType("PROPOSE", []byte{0b11111111})

		go rbc(p, data.Data, id, outputChannel, idx, mm)
	}
}

func rbc(p *party.HonestParty, data []byte, id []byte, outputChannel chan []byte, idx uint32, mm map[uint32]chan bool) {
	if isBad(p.PID, p.F) {
		data[0] = data[0] + 6
	}

	p.Broadcast(&protobuf.Message{
		Data: data,
		Type: "ECHO",
		Id:   id,
	})

	f := int(p.F)

	sendReady := false
	map_echo := make(map[string]int)
	map_ready := make(map[string]int)
	mu := new(sync.Mutex)

	var wg sync.WaitGroup
	wg.Add(2)

	// stage_ready
	go func() {
		for {
			m, _ := p.GetMessageWithType("ECHO", id)
			mu.Lock()
			if !sendReady {
				if m != nil {
					key := string(m.Data)
					map_echo[key]++
					if map_echo[key] >= 2*f+1 {
						p.Broadcast(&protobuf.Message{
							Data: m.Data,
							Type: "READY",
							Id:   id,
						})
						sendReady = true
						//fmt.Println(id, " sned READ1\t", string(m.Data))
						break
					}
				}
			} else {
				break
			}
			mu.Unlock()
		}
		mu.Unlock()
		wg.Done()
	}()

	go func() {
		for {
			m, _ := p.GetMessageWithType("READY", id)
			mu.Lock()
			if m != nil {
				key := string(m.Data)
				map_ready[key]++
				if map_ready[key] >= f+1 && !sendReady {
					p.Broadcast(&protobuf.Message{
						Data: m.Data,
						Type: "READY",
						Id:   id,
					})
					sendReady = true
					//fmt.Println(id, " sned READ2\t", string(m.Data))
					mu.Unlock()
					continue
				}

				if map_ready[key] >= 2*f+1 {

					<-mm[idx-1]

					outputChannel <- m.Data

					mm[idx] <- true
					mu.Unlock()
					break
				}
			}
			mu.Unlock()
		}

		wg.Done()
	}()

	wg.Wait()
}
