package trader

import (
	"fmt"
	"log"
	"time"

	tc "github.com/ldcicconi/trading-common"
)

type Trader struct {
	OrderbookChan chan tc.SFOXOrderbook // the Trader receives orderbooks from the TraderManager through this channel
	Pair          tc.Pair
	Config        TraderConfig
	Logger        *log.Logger
}

func NewTrader(pair tc.Pair, config TraderConfig, logger *log.Logger) *Trader {
	return &Trader{
		OrderbookChan: make(chan tc.SFOXOrderbook),
		Pair:          pair,
		Config:        config,
		Logger:        logger,
	}
}

func (t *Trader) Start() {
	t.monitorOrderbooks()
}

func (t *Trader) monitorOrderbooks() {
	go func() {
		for o := range t.OrderbookChan {
			t.handleOrderbook(o)
		}
	}()
}

func (t *Trader) infof(format string, v ...interface{}) {
	format = fmt.Sprintf("[trader-%s] ", t.Pair.String()) + format
	t.Logger.Printf(format, v...)
}

func (t *Trader) handleOrderbook(o tc.SFOXOrderbook) {
	totalTime := time.Now().Sub(o.SFOXTimestamp)
	internalLatency := time.Now().Sub(o.ReceiptTimestamp)
	networkLatency := o.ReceiptTimestamp.Sub(o.SFOXTimestamp)
	t.infof(o.DescribeArb())
	t.infof("LATENCY internal: %s network: %s total: %s", internalLatency.String(), networkLatency.String(), totalTime.String())
	arb := FindArb(o, t.Config.FeeRateBps, t.Config.ProfitThresholdBps, t.Config.MaxPositionQuantity)
	if arb != nil {
		fmt.Printf("%+v\n", *arb)
	}
}
