package main

import (
	"Dory/internal/aab"
	"Dory/internal/party"
	"Dory/pkg/config"
	"Dory/pkg/core"
	"Dory/pkg/utils/logger"
	"crypto/rand"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func main() {
	log.Println("start consensus init")
	B, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	c, err := config.NewConfig("./config.yaml", true)
	if err != nil {
		log.Fatalln(err)
	}
	logg := logger.NewLoggerWithID("config", c.PID)
	tssConfig := config.TSSconfig{}
	err = tssConfig.UnMarshal(c.TSSconfig)
	if err != nil {
		logg.Fatalf("fail to unmarshal tssConfig: %s", err.Error())
	}

	p := party.NewHonestParty(uint32(c.N), uint32(c.F), uint32(c.PID), c.IPList, c.PortList, tssConfig.Pk, tssConfig.Sk)
	p.InitReceiveChannel()

	time.Sleep(time.Second * time.Duration(c.PrepareTime))

	p.InitSendChannel()

	txlength := 250
	inputChannel := make(chan []byte, 1024)
	outputChannel := make(chan []byte, 1024)

	if B == 0 {
		B = c.N
	}

	data := make([]byte, txlength*B/c.N)
	rand.Read(data)
	go func() {
		for e := 1; e <= c.TestEpochs; e++ {
			inputChannel <- data
		}
	}()

	st := time.Now()
	resultLen := 0
	log.Println("start consensus for B=", B)
	go aab.MainProgress(p, inputChannel, outputChannel)
	for e := 1; e <= c.TestEpochs; e++ {
		value := <-outputChannel
		resultLen += len(value)
	}

	txNums := resultLen / txlength
	lantency := time.Since(st).Seconds()
	core.Mu.Lock()
	traffic := core.Traffic
	core.Mu.Unlock()

	titleList := []string{"ID", "txNums", "Lantency", "Traffic"}
	f := excelize.NewFile()
	f.SetSheetRow("Sheet1", "A1", &titleList)
	f.SetSheetRow("Sheet1", "A2", &[]interface{}{p.PID, txNums, lantency, traffic})
	if err := f.SaveAs("statistic.xlsx"); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second * time.Duration(c.WaitTime))
}
