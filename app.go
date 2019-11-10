package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"

	tc "github.com/ldcicconi/trading-common"
	ws "github.com/ldcicconi/ws-contractor"
)

type app struct {
	md            *marketData
	tm            *traderManager
	rawDataChan   chan ws.MessageEnvelope
	orderbookChan chan tc.SFOXOrderbook
}

func NewApp(wsURL url.URL, wsSubMessage interface{}, wsIsSecure bool, sfoxAPIKey string, pairsStr []string) *app {
	subMessageBytes, _ := json.Marshal(wsSubMessage)
	pairs := GetPairsFromPairStrings(pairsStr)
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	return &app{
		md:            NewMarketData(wsURL, subMessageBytes, wsIsSecure, logger),
		tm:            NewTraderManager(logger, sfoxAPIKey, pairs),
		rawDataChan:   make(chan ws.MessageEnvelope),
		orderbookChan: make(chan tc.SFOXOrderbook),
	}
}

func NewSFOXArbApp(pairsStr []string, SFOXAPIKey string) *app {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	pairs := GetPairsFromPairStrings(pairsStr)
	return &app{
		md:            NewSFOXMarketData(SFOXURL, pairs, logger),
		tm:            NewTraderManager(logger, SFOXAPIKey, pairs),
		rawDataChan:   make(chan ws.MessageEnvelope),
		orderbookChan: make(chan tc.SFOXOrderbook),
	}
}

func (a *app) Start() {
	// start the marketdata service
	a.md.Start(a.rawDataChan, a.orderbookChan)
	// start the traders
	a.tm.Start(a.orderbookChan)
}
