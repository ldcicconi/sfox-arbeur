package main

import (
	"testing"
	"time"

	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

var (
	// no arb
	testOrderbookOne = tc.SFOXOrderbook{
		Orderbook: tc.Orderbook{
			Asks: []tc.Offer{
				tc.Offer{
					Price:    decimal.New(101, 0),
					Quantity: decimal.New(10, 0),
				},
				tc.Offer{
					Price:    decimal.New(102, 0),
					Quantity: decimal.New(5, 0),
				},
				tc.Offer{
					Price:    decimal.New(103, 0),
					Quantity: decimal.New(7, 0),
				},
			},
			Bids: []tc.Offer{
				tc.Offer{
					Price:    decimal.New(99, 0),
					Quantity: decimal.New(10, 0),
				},
				tc.Offer{
					Price:    decimal.New(98, 0),
					Quantity: decimal.New(5, 0),
				},
				tc.Offer{
					Price:    decimal.New(97, 0),
					Quantity: decimal.New(6, 0),
				},
			},
		},
		SFOXTimestamp: time.Now(),
		Pair:          *tc.NewPair("btcusd"),
	}
	// simple arb - BUY at 100, SELL at 101. 100BPS arb. Max 8BTC available to buy.
	testOrderbookTwo = tc.SFOXOrderbook{
		Orderbook: tc.Orderbook{
			Asks: []tc.Offer{
				tc.Offer{
					Price:    decimal.New(100, 0),
					Quantity: decimal.New(8, 0),
				},
				tc.Offer{
					Price:    decimal.New(101, 0),
					Quantity: decimal.New(10, 0),
				},
				tc.Offer{
					Price:    decimal.New(102, 0),
					Quantity: decimal.New(5, 0),
				},
				tc.Offer{
					Price:    decimal.New(103, 0),
					Quantity: decimal.New(7, 0),
				},
			},
			Bids: []tc.Offer{
				tc.Offer{
					Price:    decimal.New(101, 0),
					Quantity: decimal.New(10, 0),
				},
				tc.Offer{
					Price:    decimal.New(98, 0),
					Quantity: decimal.New(5, 0),
				},
				tc.Offer{
					Price:    decimal.New(97, 0),
					Quantity: decimal.New(6, 0),
				},
			},
		},
		SFOXTimestamp: time.Now(),
		Pair:          *tc.NewPair("btcusd"),
	}
	// complex arb - existing on multiple bid-ask pairs
	testOrderbookThree = tc.SFOXOrderbook{
		Orderbook: tc.Orderbook{
			Asks: []tc.Offer{
				tc.Offer{
					Price:    decimal.New(99, 0),
					Quantity: decimal.New(1, 0),
				},
				tc.Offer{
					Price:    decimal.New(100, 0),
					Quantity: decimal.New(3, 0),
				},
				tc.Offer{
					Price:    decimal.New(1004, -1),
					Quantity: decimal.New(5, 0),
				},
				tc.Offer{
					Price:    decimal.New(103, 0),
					Quantity: decimal.New(7, 0),
				},
			},
			Bids: []tc.Offer{
				tc.Offer{
					Price:    decimal.New(102, 0),
					Quantity: decimal.New(2, 0),
				},
				tc.Offer{
					Price:    decimal.New(101, 0),
					Quantity: decimal.New(2, 0),
				},
				tc.Offer{
					Price:    decimal.New(1008, -1),
					Quantity: decimal.New(1, 0),
				},
				tc.Offer{
					Price:    decimal.New(10069, -2),
					Quantity: decimal.New(1, 0),
				},
				tc.Offer{
					Price:    decimal.New(98, 0),
					Quantity: decimal.New(5, 0),
				},
				tc.Offer{
					Price:    decimal.New(97, 0),
					Quantity: decimal.New(6, 0),
				},
			},
		},
		SFOXTimestamp: time.Now(),
		Pair:          *tc.NewPair("btcusd"),
	}
)

func AreArbsIdentical(arb1, arb2 arbStrat) bool {
	if !arb1.BuyPrice.Truncate(4).Equal(arb2.BuyPrice.Truncate(4)) {
		return false
	}
	if !arb1.SellPrice.Truncate(4).Equal(arb2.SellPrice.Truncate(4)) {
		return false
	}
	if !arb1.BuyLimitPrice.Truncate(4).Equal(arb2.BuyLimitPrice.Truncate(4)) {
		return false
	}
	if !arb1.SellLimitPrice.Truncate(4).Equal(arb2.SellLimitPrice.Truncate(4)) {
		return false
	}
	if !arb1.ProfitGoal.Truncate(4).Equal(arb2.ProfitGoal.Truncate(4)) {
		return false
	}
	if !arb1.ProfitGoalBps.Truncate(4).Equal(arb2.ProfitGoalBps.Truncate(4)) {
		return false
	}
	if !arb1.Quantity.Truncate(4).Equal(arb2.Quantity.Truncate(4)) {
		return false
	}
	if arb1.Pair.String() != arb2.Pair.String() {
		return false
	}
	return true
}

var testLimits = TradeLimits{
	MinOrderQuantity:   decimal.Zero,
	MaxOrderQuantity:   decimal.New(100000, 0),
	MinOrderAmount:     decimal.Zero,
	MaxOrderAmount:     decimal.New(10000000, 0),
	ProfitThresholdBps: decimal.New(30, 0),
	FeeRateBps:         decimal.New(10, 0),
}

func TestNoArbOrderbook(t *testing.T) {
	_, err := FindArb(testOrderbookOne, testLimits, decimal.New(1000, 0))
	if err != errNoArb {
		t.Errorf("FAILED TEST ON NON-EXISTANT ARB WITH NO FEES AND 1BPS PROFIT MIN")
	}
	_, err = FindArb(testOrderbookOne, testLimits, decimal.New(1000, 0))
	if err != errNoArb {
		t.Errorf("FAILED TEST ON NON-EXISTANT ARB WITH 17.5BPS FEES AND 1BPS PROFIT MIN")
	}
	_, err = FindArb(testOrderbookOne, testLimits, decimal.New(1000, 0))
	if err != errNoArb {
		t.Errorf("FAILED TEST ON NON-EXISTANT ARB WITH NO FEES AND 100BPS PROFIT MIN")
	}
}

func TestSimpleArb(t *testing.T) {
	// arb with perfect max amount, no fees, 1bps limit
	arb1, err := FindArb(testOrderbookTwo, testLimits, decimal.New(800, 0)) // 800 + no fee means we can buy the entire offer
	expectedArb1 := arbStrat{
		Pair:           *tc.NewPair("btcusd"),
		BuyPrice:       decimal.New(1001, -1),
		SellPrice:      decimal.New(101101, -3),
		BuyLimitPrice:  decimal.New(100, 0),
		SellLimitPrice: decimal.New(101, 0),
		Quantity:       decimal.New(7992, -3),
		ProfitGoal:     decimal.New(79999, -4), // ~$1 on each btc purchased
		ProfitGoalBps:  decimal.New(100, 0),
	}
	if err == errNoArb {
		t.Errorf("FAILED TEST ON EXISTANT ARB WITH NO FEES AND 1BPS PROFIT MIN - arb: %+v", arb1)
	}
	if !AreArbsIdentical(arb1, expectedArb1) {
		t.Errorf("FAILED TEST ON EXISTANT ARB WITH NO FEES AND 1BPS PROFIT MIN - arb: %+v", arb1)
	}
	// arb with short balance, no fees, 1bps limit
	arb2, err := FindArb(testOrderbookTwo, testLimits, decimal.New(500, 0))
	expectedArb2 := arbStrat{
		Pair:           *tc.NewPair("btcusd"),
		BuyPrice:       decimal.New(1001, -1),
		SellPrice:      decimal.New(101101, -3),
		BuyLimitPrice:  decimal.New(100, 0),
		SellLimitPrice: decimal.New(101, 0),
		Quantity:       decimal.New(4995, -3),
		ProfitGoal:     decimal.New(49999, -4), // $1 on each btc purchased
		ProfitGoalBps:  decimal.New(100, 0),
	}
	if err == errNoArb {
		t.Errorf("FAILED TEST ON EXISTANT ARB WITH NO FEES AND 1BPS PROFIT MIN - arb: %+v", arb2)
	}
	if !AreArbsIdentical(arb2, expectedArb2) {
		t.Errorf("FAILED TEST ON EXISTANT ARB WITH NO FEES AND 1BPS PROFIT MIN - arb: %+v", arb2)
	}
	// arb with short balance, 20BPS fee, 1bps limit, $400 available
	// arb3, err := FindArb(testOrderbookTwo, testLimits, decimal.New(400, 0))
	// expectedArb3 := arbStrat{
	// 	Pair:           *tc.NewPair("btcusd"),
	// 	BuyPrice:       decimal.New(1002, -1),
	// 	SellPrice:      decimal.New(100798, -3),
	// 	BuyLimitPrice:  decimal.New(100, 0),
	// 	SellLimitPrice: decimal.New(101, 0),
	// 	Quantity:       decimal.New(3992, -3),   // based on the adjusted buy price (can buy 3.992btc)
	// 	ProfitGoal:     decimal.New(23872, -4),  // $2.3872 profit
	// 	ProfitGoalBps:  decimal.New(596806, -4), // 59.6806 BPS
	// }
	// if err == errNoArb {
	// 	t.Errorf("FAILED TEST ON EXISTANT ARB WITH 20BPS FEES AND 1BPS PROFIT MIN - arb: %+v", arb3)
	// }
	// if !AreArbsIdentical(arb3, expectedArb3) {
	// 	t.Errorf("FAILED TEST ON EXISTANT ARB WITH 20BPS FEES AND 1BPS PROFIT MIN - arb: %+v", arb3)
	// }
}

// This isn't perfect, but seems to work pretty well
func TestComplexArb(t *testing.T) {
	// arb with perfect max amount, 10BPS fee, 30bps limit
	arb1, err := FindArb(testOrderbookThree, testLimits, decimal.New(800, 0))
	expectedArb1 := arbStrat{
		Pair:           *tc.NewPair("btcusd"),
		BuyPrice:       decimal.New(9997988, -5),
		SellPrice:      decimal.New(10129896, -5),
		BuyLimitPrice:  decimal.New(1004, -1),
		SellLimitPrice: decimal.New(1008, -1),
		Quantity:       decimal.New(5, 0),
		ProfitGoal:     decimal.New(65954, -4),
		ProfitGoalBps:  decimal.New(13193454, -5),
	}
	if err == errNoArb {
		t.Errorf("FAILED TEST ON EXISTANT ARB WITH NO FEES AND 1BPS PROFIT MIN - arb: %+v", arb1)
	}
	if !AreArbsIdentical(arb1, expectedArb1) {
		t.Errorf("FAILED TEST ON EXISTANT ARB WITH NO FEES AND 1BPS PROFIT MIN - arb: %+v", arb1)
	}
}
