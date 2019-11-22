package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	sfoxapi "github.com/ldcicconi/sfox-api-lib"
	tc "github.com/ldcicconi/trading-common"
)

type Trader struct {
	OrderbookChan       chan tc.SFOXOrderbook // the Trader receives orderbooks from the TraderManager through this channel
	Config              TraderConfig
	Logger              *log.Logger
	manager             *traderManager
	currentPosition     *arbStrat
	killChan            chan bool // the arbMonitor loop listens on this, and will exit the position if signalled
	buyOrderStatus      *sfoxapi.OrderStatusResponse
	buyOrderStatusChan  chan bool // a goroutine notifies the main arbMonitor of buy order updates through this chan
	sellOrderStatus     *sfoxapi.OrderStatusResponse
	sellOrderStatusChan chan bool
}

func NewTrader(config TraderConfig, logger *log.Logger, manager *traderManager) *Trader {
	return &Trader{
		OrderbookChan:       make(chan tc.SFOXOrderbook),
		Config:              config,
		Logger:              logger,
		manager:             manager,
		killChan:            make(chan bool),
		buyOrderStatusChan:  make(chan bool),
		sellOrderStatusChan: make(chan bool),
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
	format = fmt.Sprintf("[trader-%s] ", t.Config.Pair.String()) + format
	t.Logger.Printf(format, v...)
}

func (t *Trader) logLatency(ob tc.SFOXOrderbook) {
	totalTime := time.Now().Sub(ob.SFOXTimestamp)
	internalLatency := time.Now().Sub(ob.ReceiptTimestamp)
	networkLatency := ob.ReceiptTimestamp.Sub(ob.SFOXTimestamp)
	t.infof("LATENCY internal: %s network: %s total: %s", internalLatency.String(), networkLatency.String(), totalTime.String())
}

func (t *Trader) handleOrderbook(o tc.SFOXOrderbook) {
	quoteBalance := t.getBalance(t.Config.Pair.Quote)
	maxPosAmount := decimal.Min(quoteBalance, t.Config.MaxPositionAmount)
	arb, err := FindArb(o, t.Config.FeeRateBps, t.Config.ProfitThresholdBps, maxPosAmount)
	if err == errNoArb && t.currentPosition == nil {
		// do nothing - there is no arb, and we do not need to try to cut losses.
		return
	}
	if err == errNoArb && t.currentPosition != nil {
		// we might want to cut losses - think about this later
		// t.killChan <- true
		return
	}
	if err == nil && t.currentPosition == nil {
		// enter into a position, as there is a profitable arb opportunity, per our parameters
		t.infof("found a profitable arb: %+v", arb)
		t.currentPosition = &arb
		t.manageArbStrategy()
		return
	}
	if err == nil && t.currentPosition != nil {
		// do nothing - we are in the middle of trying to exit an arb.
		return
	}

}

func (t *Trader) manageArbStrategy() {
	go func() {
		for {
			select {
			case <-t.killChan:
				// exit the position
				if t.currentPosition.Status == STATUS_BUY_STARTED {
					t.cancelOrder(t.buyOrderStatus.ID)
				} else if t.currentPosition.Status == STATUS_SELL_STARTED {
					t.cancelOrder(t.sellOrderStatus.ID)
				}
				return
			case <-t.buyOrderStatusChan:
				fmt.Println("update from buy order status channel")
				// update fill information if anything has changed
				if t.buyOrderStatus.FilledQuantity.Equal(t.currentPosition.Quantity) {
					// complete fill:
					t.infof("[buy] RECOGNIZED TOTAL FILL. FILLEDQUANTITY: ", t.buyOrderStatus.FilledQuantity.String())
					t.currentPosition.Status = STATUS_BUY_COMPLETE
				} else {
					t.infof("[buy] RECOGNIZED PARTIAL FILL. FILLEDQUANTITY: ", t.buyOrderStatus.FilledQuantity.String())
				}
			case <-t.sellOrderStatusChan:
				// update fill information if anything has changed
				if t.sellOrderStatus.FilledQuantity.Equal(t.currentPosition.Quantity) {
					// complete fill:
					t.infof("[sell] RECOGNIZED TOTAL FILL. FILLEDQUANTITY: ", t.sellOrderStatus.FilledQuantity.String())
					t.currentPosition.Status = STATUS_SELL_COMPLETE
				} else {
					t.infof("[sell] RECOGNIZED PARTIAL FILL. FILLEDQUANTITY: ", t.sellOrderStatus.FilledQuantity.String())
				}
			default:
			}
			if t.currentPosition.Status == STATUS_INIT {
				// enter the position
				buyOrder := NewBuyOrderFromArbStrat(*t.currentPosition)
				t.infof("attempting to buy %+v", buyOrder)
				status, err := t.executeOrder(*buyOrder)
				if err != nil {
					t.infof("error attempting to buy %s", err.Error())
					continue
				}
				t.infof("buy request successful!")
				// determine status
				statusLower := strings.ToLower(status.Status)
				if statusLower == "started" {
					t.infof("buy started")
					t.currentPosition.Status = STATUS_BUY_STARTED
					t.currentPosition.BuyTime = time.Now()
					t.startBuyOrderStatusLoop(status.ID)
				} else {
					t.infof("unrecognized status: %s", statusLower)
				}
				continue
			}
			if t.currentPosition.Status == STATUS_BUY_COMPLETE {
				// exit the position
				sellOrder := NewSellOrderFromArbStrat(*t.currentPosition, t.buyOrderStatus.FilledQuantity)
				status, err := t.executeOrder(*sellOrder)
				if err != nil {
					continue
				}
				// determine status
				statusLower := strings.ToLower(status.Status)
				if statusLower == "started" {
					t.currentPosition.Status = STATUS_SELL_STARTED
					t.currentPosition.BuyTime = time.Now()
					t.startBuyOrderStatusLoop(status.ID)
				} else {
					t.infof("unrecognized status: %s", statusLower)
				}
				continue
			}
			if t.currentPosition.Status == STATUS_SELL_COMPLETE {
				t.infof("ARB COMPLETE. PROFIT: %s%s", t.buyOrderStatus.FilledQuantity.Mul(t.buyOrderStatus.VWAP).Sub(t.sellOrderStatus.FilledQuantity.Mul(t.sellOrderStatus.VWAP)).String(), string(t.Config.Pair.Quote))
				return
			}
			if t.currentPosition.Status == STATUS_BUY_STARTED && time.Now().Sub(t.currentPosition.BuyTime).Seconds() > 8.0 {
				// cancel if it's taking too long to fill our buy order
				t.cancelOrder(t.buyOrderStatus.ID)
				return
			}
		}
	}()
}

func (t *Trader) startBuyOrderStatusLoop(orderID int64) {
	t.infof("starting buy order status loop for %v", orderID)
	go func() {
		for {
			time.Sleep(time.Millisecond * 500)
			status, err := t.getOrderStatus(orderID)
			if err != nil {
				t.infof("ERROR: %s", err.Error())
				continue
			}
			if t.buyOrderStatus == nil {
				t.buyOrderStatus = &status
				t.buyOrderStatusChan <- true //notify the loop that there was an order status update
				continue
			}
			if status.FilledQuantity.GreaterThan(t.buyOrderStatus.FilledQuantity) {
				t.buyOrderStatus = &status
				t.buyOrderStatusChan <- true //notify the loop that there was an order status update
				continue
			} else if t.buyOrderStatus.Status == "done" {
				t.buyOrderStatus = &status
				t.buyOrderStatusChan <- true //notify the loop that there was an order status update
				return
			}
		}
	}()
}

func (t *Trader) startSellOrderStatusLoop(orderID int64) {
	t.infof("starting sell order status loop for %v", orderID)
	go func() {
		for {
			time.Sleep(time.Millisecond * 500)
			status, err := t.getOrderStatus(orderID)
			if err != nil {
				t.infof("ERROR: %s", err.Error())
				continue
			}
			if t.sellOrderStatus == nil {
				t.sellOrderStatus = &status
				t.sellOrderStatusChan <- true //notify the loop that there was an order status update
				continue
			}
			if status.FilledQuantity.GreaterThan(t.sellOrderStatus.FilledQuantity) {
				t.sellOrderStatus = &status
				t.sellOrderStatusChan <- true //notify the loop that there was an order status update
				continue
			} else if t.sellOrderStatus.Status == "done" {
				t.sellOrderStatus = &status
				t.sellOrderStatusChan <- true //notify the loop that there was an order status update
				return
			}
		}
	}()
}

func (t *Trader) executeOrder(orderParams TraderOrder) (orderStatus sfoxapi.OrderStatusResponse, err error) {
	client := t.manager.GetSFOXClient()
	orderStatus, err = client.NewOrder(orderParams.Quantity, orderParams.LimitPrice, orderParams.AlgoID, orderParams.Pair.String(), string(orderParams.Side))
	return
}

func (t *Trader) getOrderStatus(id int64) (orderStatus sfoxapi.OrderStatusResponse, err error) {
	client := t.manager.GetSFOXClient()
	orderStatus, err = client.OrderStatus(id)
	return
}

func (t *Trader) cancelOrder(id int64) (err error) {
	client := t.manager.GetSFOXClient()
	err = client.CancelOrder(id)
	return
}

func (t *Trader) getBalance(c tc.Currency) decimal.Decimal {
	return t.manager.GetBalance(c)
}
