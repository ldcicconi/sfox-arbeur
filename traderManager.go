package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	sfox "github.com/ldcicconi/sfox-api-lib"
	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

type SafeBalanceMap struct {
	mtx sync.RWMutex
	m   map[tc.Currency]decimal.Decimal
}

func NewSafeBalanceMap() *SafeBalanceMap {
	return &SafeBalanceMap{
		m: make(map[tc.Currency]decimal.Decimal),
	}
}

type traderManager struct {
	Logger         *log.Logger
	SFOXClientPool *SFOXAPIClientPool
	balances       *SafeBalanceMap
	traders        map[tc.Pair]*Trader // one trader per pair
}

func NewTraderManager(logger *log.Logger, sfoxAPIKeys []string, traderConfigs []TraderConfig) *traderManager {
	traders := make(map[tc.Pair]*Trader)
	for _, tc := range traderConfigs {
		traders[tc.Pair] = NewTrader(tc, logger, nil)
	}
	return &traderManager{
		Logger:         logger,
		balances:       NewSafeBalanceMap(),
		SFOXClientPool: NewSFOXAPIClientPool(sfoxAPIKeys),
		traders:        traders,
	}
}

func (t *traderManager) LogInfo(text string) {
	t.Logger.Println("[traderManager] [info] " + text)
}

func (t *traderManager) Start(orderbookChan chan tc.SFOXOrderbook) {
	t.initTraders()
	t.monitorBalances()
	time.Sleep(2 * time.Second)
	t.startTraders()
	t.routeOrderbooks(orderbookChan)
}

func (tm *traderManager) initTraders() {
	for _, t := range tm.traders {
		t.manager = tm
	}
}

func (t *traderManager) startTraders() {
	for _, trader := range t.traders {
		trader.Start()
	}
}

func (t *traderManager) routeOrderbooks(orderbookChan chan tc.SFOXOrderbook) {
	go func() {
		for o := range orderbookChan {
			t.traders[o.Pair].OrderbookChan <- o
		}
	}()
}

func (t *traderManager) monitorBalances() {
	// Poll SFOX every 5 seconds and update local register
	go func() {
		var balances []sfox.SFOXBalance
		var err error
		for range time.Tick(5 * time.Second) {
			client := t.SFOXClientPool.GetAPIClient()
			balances, err = client.GetBalances()
			if err != nil {
				t.Logger.Printf("error getting balances %s", err.Error())
				continue
			}
			t.SFOXClientPool.ReturnAPIClient(client)
			t.balances.mtx.Lock()
			for _, b := range balances {
				t.balances.m[tc.Currency(b.Currency)] = b.Available
			}
			t.balances.mtx.Unlock()
		}
	}()
}

func (t *traderManager) logArb(o tc.SFOXOrderbook) {
	arb := o.Arb()
	arbBps := o.ArbBps()
	condition := "arb"
	if arb.LessThanOrEqual(decimal.Zero) {
		condition = "spread"
		arb = arb.Neg()
		arbBps = arbBps.Neg()
	}
	t.LogInfo(fmt.Sprintf("%s %s: %s (%s bps)", o.Pair, condition, arb, arbBps))
}

func (tm *traderManager) GetSFOXClient() *sfox.SFOXAPI {
	return tm.SFOXClientPool.GetAPIClient()
}

func (tm *traderManager) ReturnSFOXClient(c *sfox.SFOXAPI) {
	tm.SFOXClientPool.ReturnAPIClient(c)
}

func (tm *traderManager) GetBalance(c tc.Currency) (balance decimal.Decimal) {
	tm.balances.mtx.RLock()
	defer tm.balances.mtx.RUnlock()
	return tm.balances.m[c]
}
