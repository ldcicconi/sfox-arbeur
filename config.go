package main

import (
	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

type TraderConfig struct {
	Pair               tc.Pair
	MaxPositionAmount  decimal.Decimal // in quote currency of the pair
	ProfitThresholdBps decimal.Decimal
	FeeRateBps         decimal.Decimal
}

func NewTraderConfig(pair tc.Pair, maxPosAmount, profitThreshold, feeRate decimal.Decimal) *TraderConfig {
	return &TraderConfig{
		Pair:               pair,
		MaxPositionAmount:  maxPosAmount,
		ProfitThresholdBps: profitThreshold,
		FeeRateBps:         feeRate,
	}
}
