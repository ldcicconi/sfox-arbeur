package main

import (
	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

type TraderOrder struct {
	Side       tc.Side
	Pair       tc.Pair
	Quantity   decimal.Decimal
	LimitPrice decimal.Decimal
	AlgoID     int
}

func NewBuyOrderFromArbStrat(arb arbStrat) *TraderOrder {
	return &TraderOrder{
		Side:       tc.SIDE_BUY,
		Pair:       arb.Pair,
		Quantity:   arb.Quantity,
		LimitPrice: arb.BuyLimitPrice,
		AlgoID:     200,
	}
}

func NewSellOrderFromArbStrat(arb arbStrat, quantity decimal.Decimal) *TraderOrder {
	return &TraderOrder{
		Side:       tc.SIDE_SELL,
		Pair:       arb.Pair,
		Quantity:   quantity,
		LimitPrice: arb.SellLimitPrice,
		AlgoID:     200,
	}
}
