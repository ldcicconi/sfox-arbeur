package main

import (
	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

type TraderConfig struct {
	Pair tc.Pair
	TradeLimits
}

type TradeLimits struct {
	MinOrderQuantity   decimal.Decimal // in base currency
	MaxOrderQuantity   decimal.Decimal // ''
	MinOrderAmount     decimal.Decimal // in quote currency
	MaxOrderAmount     decimal.Decimal // ''
	ProfitThresholdBps decimal.Decimal
	FeeRateBps         decimal.Decimal
}

func NewTraderConfig(pair tc.Pair, limits TradeLimits) *TraderConfig {
	return &TraderConfig{
		Pair:        pair,
		TradeLimits: limits,
	}
}
