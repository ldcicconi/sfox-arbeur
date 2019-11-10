package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	tc "github.com/ldcicconi/trading-common"
	ws "github.com/ldcicconi/ws-contractor"
)

type marketData struct {
	wsWorker *ws.WsContractor
	Logger   *log.Logger
}

func NewMarketData(marketURL url.URL, subMessage []byte, isSecure bool, logger *log.Logger) *marketData {
	ws := ws.NewWsContractor(marketURL, subMessage, isSecure)
	return &marketData{
		wsWorker: ws,
		Logger:   logger,
	}
}

func NewSFOXMarketData(marketURL url.URL, pairs []tc.Pair, logger *log.Logger) *marketData {
	sfoxSubMessage := GenerateSFOXOrderbookSubMessage(pairs)
	fmt.Println(sfoxSubMessage)
	bodyBytes, _ := json.Marshal(sfoxSubMessage)
	fmt.Println(string(bodyBytes))
	ws := ws.NewWsContractor(marketURL, bodyBytes, true)
	return &marketData{
		wsWorker: ws,
		Logger:   logger,
	}
}

func (md *marketData) Start(rawDataChan chan ws.MessageEnvelope, orderbookChan chan tc.SFOXOrderbook) {
	md.wsWorker.Consume(rawDataChan)
	md.ProcessData(rawDataChan, orderbookChan)
}

func (md *marketData) LogInfo(text string) {
	md.Logger.Println("[marketdata] [info] " + text)
}

func (md *marketData) ProcessData(rawDataChan chan ws.MessageEnvelope, orderbookChan chan tc.SFOXOrderbook) {
	go func() {
		for msg := range rawDataChan {
			// md.LogInfo("unmarshalling json")
			o, err := tc.NewSFOXOrderbookFromJSON(msg.Payload, msg.ReceiptTimestamp)
			// md.LogInfo("unmarshalling json complete")
			if err == tc.ErrFirstMessage {
				md.Logger.Println("Received messsage w/ sequence=1")
				continue
			} else if err != nil {
				md.Logger.Println("ERROR: " + err.Error())
				continue
			}
			orderbookChan <- *o
		}
	}()
}
