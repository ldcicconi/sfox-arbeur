package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	sfox "github.com/ldcicconi/sfox-api-lib"
	"github.com/ldcicconi/sfox-arbeur/trader"
	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

type SafeBalanceMap struct {
	mtx sync.RWMutex
	m   map[Currency]decimal.Decimal
}

func NewSafeBalanceMap() *SafeBalanceMap {
	return &SafeBalanceMap{
		m: make(map[Currency]decimal.Decimal),
	}
}

type traderManager struct {
	Logger   *log.Logger
	SFOX     *sfox.SFOXAPI
	balances *SafeBalanceMap
	pairs    []tc.Pair
	traders  map[tc.Pair]*trader.Trader // one trader per pair
}

func NewTraderManager(logger *log.Logger, sfoxAPIKey string, pairs []tc.Pair) *traderManager {
	return &traderManager{
		Logger:   logger,
		balances: NewSafeBalanceMap(),
		SFOX:     sfox.NewSFOXAPI(sfoxAPIKey),
		pairs:    pairs,
	}
}

func (t *traderManager) LogInfo(text string) {
	t.Logger.Println("[traderManager] [info] " + text)
}

func (t *traderManager) Start(orderbookChan chan tc.SFOXOrderbook) {
	t.initTraders()
	t.startTraders()
	t.monitorBalances()
	t.routeOrderbooks(orderbookChan)
}

func (t *traderManager) initTraders() {
	t.traders = make(map[tc.Pair]*trader.Trader)
	for _, pair := range t.pairs {
		t.traders[pair] = trader.NewTrader(pair, *t.buildTraderConfig(pair), t.Logger)
	}
}

func (t *traderManager) startTraders() {
	for _, trader := range t.traders {
		trader.Start()
	}
}

func (t *traderManager) buildTraderConfig(pair tc.Pair) *trader.TraderConfig {
	return &trader.TraderConfig{}
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
			t.LogInfo("updating balances")
			balances, err = t.SFOX.GetBalances()
			if err != nil {
				t.Logger.Printf("error getting balances %s", err.Error())
				continue
			}
			t.balances.mtx.Lock()
			for _, b := range balances {
				t.balances.m[Currency(b.Currency)] = b.Available
			}
			t.balances.mtx.Unlock()
			t.LogInfo(fmt.Sprintf("updated balances: %+v", t.balances))
		}
	}()
}

func (t *traderManager) logArb(o *tc.SFOXOrderbook) {
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
