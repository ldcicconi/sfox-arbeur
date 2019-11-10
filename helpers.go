package main

import (
	"fmt"

	tc "github.com/ldcicconi/trading-common"
)

type SFOXOrderbookSubMessage struct {
	Type  string   `json:"type"`
	Feeds []string `json:"feeds"`
}

func GenerateSFOXOrderbookSubMessage(pairs []tc.Pair) *SFOXOrderbookSubMessage {
	var feeds []string
	for _, pair := range pairs {
		feeds = append(feeds, fmt.Sprintf("orderbook.sfox.%s", pair.String()))
	}
	return &SFOXOrderbookSubMessage{
		Type:  "subscribe",
		Feeds: feeds,
	}
}

func GetPairsFromPairStrings(pairs []string) (ret []tc.Pair) {
	for _, p := range pairs {
		ret = append(ret, *tc.NewPair(p))
	}
	return ret
}
