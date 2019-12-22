package main

import (
	"fmt"
	"log"
	"strings"
	"sync/atomic"
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
	errCount            int
	arbLocker           uint32                           // ensure that trader only runs one arb at a time
	killChan            chan bool                        // the arbMonitor loop listens on this, and will exit the position if signalled
	buyOrderStatusChan  chan sfoxapi.OrderStatusResponse // a goroutine notifies the main arbMonitor of buy order updates through this chan
	sellOrderStatusChan chan sfoxapi.OrderStatusResponse
}

func NewTrader(config TraderConfig, logger *log.Logger, manager *traderManager) *Trader {
	return &Trader{
		OrderbookChan:       make(chan tc.SFOXOrderbook),
		Config:              config,
		Logger:              logger,
		manager:             manager,
		killChan:            make(chan bool),
		buyOrderStatusChan:  make(chan sfoxapi.OrderStatusResponse),
		sellOrderStatusChan: make(chan sfoxapi.OrderStatusResponse),
	}
}

func (t *Trader) Start() {
	t.monitorOrderbooks()
}

func (t *Trader) monitorOrderbooks() {
	go func() {
		for o := range t.OrderbookChan {
			t.infof(o.DescribeArb(t.Config.FeeRateBps))
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

func (t *Trader) arbActive() bool {
	if !atomic.CompareAndSwapUint32(&t.arbLocker, 0, 1) {
		return true
	}
	return false
}

func (t *Trader) handleOrderbook(o tc.SFOXOrderbook) {
	quoteBalance := t.getBalance(t.Config.Pair.Quote)
	arb, err := FindArb(o, t.Config.TradeLimits, quoteBalance)
	// t.infof(o.DescribeArb(t.Config.FeeRateBps))
	if err == errNoArb && !t.arbActive() {
		// do nothing - there is no arb, and we do not need to try to cut losses as we are not in a position
		return
	}
	if err == errNoArb && t.arbActive() {
		// we might want to cut losses - think about this later
		// t.killChan <- true
		return
	}
	if err == nil && !t.arbActive() {
		// enter into a position, as there is a profitable arb opportunity, per our parameters
		t.infof("found a profitable arb: %+v", arb)
		t.manageArbStrategy(arb)
		return
	}
	if err == nil && t.arbActive() {
		// do nothing - we are in the middle of trying to exit an arb.
		return
	}
}

func (t *Trader) manageArbStrategy(arb arbStrat) {
	go func() {
		t.errCount = 0
		defer atomic.StoreUint32(&t.arbLocker, 0)
		var buyOrderStatus sfoxapi.OrderStatusResponse
		var sellOrderStatus sfoxapi.OrderStatusResponse
		killChan := make(chan struct{})
		for {
			// blocking receive
			arb := <-t.arbChan
			t.errCount = 0
			var buyOrderStatus sfoxapi.OrderStatusResponse
			var sellOrderStatus sfoxapi.OrderStatusResponse
			for {
				// non-blocking
				select {
				case <-t.killChan:
					// exit the position
					if arb.Status == STATUS_BUY_STARTED {
						t.cancelOrder(buyOrderStatus.ID)
						subProcessKillChan <- struct{}{}
					} else if arb.Status == STATUS_SELL_STARTED {
						t.cancelOrder(sellOrderStatus.ID)
						subProcessKillChan <- struct{}{}
					}
					return
				case <-t.noArbChan:
					// exit the position
					if arb.Status == STATUS_BUY_STARTED {
						t.cancelOrder(buyOrderStatus.ID)
						subProcessKillChan <- struct{}{}
					}
					// leave the sell order open to attempt to exit the position still....
					break
				case buyOrderStatus = <-t.buyOrderStatusChan:
					fmt.Println("update from buy order status channel")
					// update fill information if anything has changed
					if buyOrderStatus.FilledQuantity.Equal(arb.Quantity) {
						// complete fill:
						t.infof("[buy] RECOGNIZED TOTAL FILL. FILLEDQUANTITY: %s", buyOrderStatus.FilledQuantity.String())
						arb.Status = STATUS_BUY_COMPLETE
					} else {
						t.infof("[buy] RECOGNIZED PARTIAL FILL. FILLEDQUANTITY: %s", buyOrderStatus.FilledQuantity.String())
						arb.Status = STATUS_BUY_STARTED
					}
				case sellOrderStatus = <-t.sellOrderStatusChan:
					// update fill information if anything has changed
					if sellOrderStatus.FilledQuantity.Equal(arb.Quantity) {
						// complete fill:
						t.infof("[sell] RECOGNIZED TOTAL FILL. FILLEDQUANTITY: %s", sellOrderStatus.FilledQuantity.String())
						arb.Status = STATUS_SELL_COMPLETE
					} else {
						t.infof("[sell] RECOGNIZED PARTIAL FILL. FILLEDQUANTITY: %s", sellOrderStatus.FilledQuantity.String())
						arb.Status = STATUS_SELL_STARTED
					}
				default:
				}
				if t.errCount > 5 {
					break
				}
				if arb.Status == STATUS_INIT {
					// enter the position
					buyOrder := NewBuyOrderFromArbStrat(arb)
					t.infof("attempting to buy %+v", buyOrder)
					status, err := t.executeOrder(*buyOrder)
					if err != nil {
						t.infof("error attempting to buy %s", err.Error())
						t.errCount++
						continue
					}
					t.infof("buy request successful!")
					// determine status
					statusLower := strings.ToLower(status.Status)
					if statusLower == "started" {
						t.infof("buy started")
						arb.Status = STATUS_BUY_STARTED
						arb.BuyTime = time.Now()
						t.startOrderStatusLoop(status.ID, t.buyOrderStatusChan, subProcessKillChan)
					} else {
						t.infof("unrecognized status: %s", statusLower)
					}
				}
				if arb.Status == STATUS_BUY_COMPLETE {
					t.errCount = 0
					// exit the position
					t.infof("attempting to exit position")
					sellOrder := NewSellOrderFromArbStrat(arb, buyOrderStatus.FilledQuantity)
					status, err := t.executeOrder(*sellOrder)
					if err != nil {
						t.infof("error attempting to sell %s", err.Error())
						t.errCount++
						continue
					}
					// determine status
					statusLower := strings.ToLower(status.Status)
					if statusLower == "started" {
						t.infof("sell started")
						arb.Status = STATUS_SELL_STARTED
						arb.BuyTime = time.Now()
						t.startOrderStatusLoop(status.ID, t.sellOrderStatusChan, subProcessKillChan)
					} else {
						t.infof("order %v requires manual intervention - returned status %v", status.ID, statusLower)
					}

				}
				if arb.Status == STATUS_SELL_COMPLETE {
					t.infof("ARB COMPLETE. PROFIT: %s%s", buyOrderStatus.NetProceeds.Add(sellOrderStatus.NetProceeds).String(), string(t.Config.Pair.Quote))
					break
				}
				if arb.Status == STATUS_BUY_STARTED && time.Now().Sub(arb.BuyTime).Seconds() > 8.0 {
					// cancel if it's taking too long to fill our buy order
					t.cancelOrder(buyOrderStatus.ID)
					break
				}
			}
			if t.errCount > 5 {
				return
			}
			if arb.Status == STATUS_INIT {
				// enter the position
				buyOrder := NewBuyOrderFromArbStrat(arb)
				t.infof("attempting to buy %+v", buyOrder)
				status, err := t.executeOrder(*buyOrder)
				if err != nil {
					t.infof("error attempting to buy %s", err.Error())
					t.errCount++
					continue
				}
				t.infof("buy request successful!")
				// determine status
				statusLower := strings.ToLower(status.Status)
				if statusLower == "started" {
					t.infof("buy started")
					arb.Status = STATUS_BUY_STARTED
					arb.BuyTime = time.Now()
					t.startOrderStatusLoop(status.ID, t.buyOrderStatusChan, killChan)
				} else {
					t.infof("unrecognized status: %s", statusLower)
				}
			}
			if arb.Status == STATUS_BUY_COMPLETE {
				t.errCount = 0
				// exit the position
				t.infof("attempting to exit position")
				sellOrder := NewSellOrderFromArbStrat(arb, buyOrderStatus.FilledQuantity)
				status, err := t.executeOrder(*sellOrder)
				if err != nil {
					t.infof("error attempting to sell %s", err.Error())
					t.errCount++
					continue
				}
				// determine status
				statusLower := strings.ToLower(status.Status)
				if statusLower == "started" {
					t.infof("sell started")
					arb.Status = STATUS_SELL_STARTED
					arb.BuyTime = time.Now()
					t.startOrderStatusLoop(status.ID, t.sellOrderStatusChan, killChan)
				} else {
					t.infof("order %v requires manual intervention - returned status %v", status.ID, statusLower)
				}

			}
			if arb.Status == STATUS_SELL_COMPLETE {
				t.infof("ARB COMPLETE. PROFIT: %s%s", buyOrderStatus.FilledQuantity.Mul(buyOrderStatus.VWAP).Sub(sellOrderStatus.FilledQuantity.Mul(sellOrderStatus.VWAP)).String(), string(t.Config.Pair.Quote))
				return
			}
			if arb.Status == STATUS_BUY_STARTED && time.Now().Sub(arb.BuyTime).Seconds() > 8.0 {
				// cancel if it's taking too long to fill our buy order
				t.cancelOrder(buyOrderStatus.ID)
				return
			}
		}
	}()
}

func (t *Trader) startOrderStatusLoop(orderID int64, statusChannel chan sfoxapi.OrderStatusResponse, killChan chan struct{}) {
	t.infof("starting buy order status loop for %v", orderID)
	var lastOrderStatus sfoxapi.OrderStatusResponse
	go func() {
		for {
			select {
			case <-killChan:
				return
			default:
			}
			time.Sleep(time.Millisecond * 500)
			newOrderStatus, err := t.getOrderStatus(orderID)
			if err != nil {
				t.infof("ERROR: %s", err.Error())
				continue
			}
			if newOrderStatus.Status == "Canceled" {
				return
			}
			if newOrderStatus.Status == "Done" {
				statusChannel <- newOrderStatus //notify the loop that there was an order status update
				return
			}
			if newOrderStatus.FilledQuantity.GreaterThan(lastOrderStatus.FilledQuantity) {
				statusChannel <- newOrderStatus //notify the loop that there was an order status update
				continue
			}
			lastOrderStatus = newOrderStatus
		}
	}()
}

func (t *Trader) executeOrder(orderParams TraderOrder) (orderStatus sfoxapi.OrderStatusResponse, err error) {
	client := t.manager.GetSFOXClient()
	orderStatus, err = client.NewOrder(orderParams.Quantity, orderParams.LimitPrice, orderParams.AlgoID, orderParams.Pair.String(), string(orderParams.Side))
	t.manager.ReturnSFOXClient(client)
	return
}

func (t *Trader) getOrderStatus(id int64) (orderStatus sfoxapi.OrderStatusResponse, err error) {
	client := t.manager.GetSFOXClient()
	orderStatus, err = client.OrderStatus(id)
	t.manager.ReturnSFOXClient(client)
	return
}

func (t *Trader) cancelOrder(id int64) (err error) {
	client := t.manager.GetSFOXClient()
	err = client.CancelOrder(id)
	t.manager.ReturnSFOXClient(client)
	return
}

func (t *Trader) getBalance(c tc.Currency) decimal.Decimal {
	return t.manager.GetBalance(c)
}
