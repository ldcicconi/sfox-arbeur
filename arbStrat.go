package main

import (
	"fmt"
	"time"

	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

var errNoArb = fmt.Errorf("there was no arb")

type SFOXSmartOrderRequest struct {
	Price decimal.Decimal
	Side  tc.Side
}

type arbStrat struct {
	Pair           tc.Pair
	BuyPrice       decimal.Decimal
	SellPrice      decimal.Decimal
	BuyLimitPrice  decimal.Decimal
	SellLimitPrice decimal.Decimal
	Quantity       decimal.Decimal // denominated in the base currency
	ProfitGoal     decimal.Decimal // denominated in the quote currency
	ProfitGoalBps  decimal.Decimal // ROI*1e5
	Status         arbStatus
	BuyTime        time.Time // the time that the trader started the buy at
}

type arbStatus int

func (as *arbStatus) String() string {
	return string(int(*as))
}

const (
	STATUS_INIT          arbStatus = 0
	STATUS_BUY_STARTED   arbStatus = 50
	STATUS_BUY_COMPLETE  arbStatus = 100
	STATUS_SELL_STARTED  arbStatus = 150
	STATUS_SELL_COMPLETE arbStatus = 200
	STATUS_DONE          arbStatus = 300
	STATUS_CANCELED      arbStatus = -1
)

// TODO: needs to be tested
func FindArb(inOb tc.SFOXOrderbook, limits TradeLimits, availableQuoteBalance decimal.Decimal) (arb arbStrat, err error) {
	o := inOb.MakeCopy()
	priceArb := o.Arb()
	if priceArb.LessThanOrEqual(decimal.Zero) {
		err = errNoArb
		return
	}
	var bidIndex int
	var cumulativeQuantitySold decimal.Decimal
	var cumulativeProceedsFromSale decimal.Decimal
	var cumulativeProceedsFromSaleWFees decimal.Decimal
	var cumulativeQuantityBought decimal.Decimal
	var cumulativeBuyCost decimal.Decimal
	var cumulativeBuyCostWFees decimal.Decimal
	var highestBuyPrice decimal.Decimal
	var lowestSellPrice decimal.Decimal
	remainingAvailableQuote := decimal.Min(limits.MaxOrderAmount, availableQuoteBalance)

	var bidSliceQuantity decimal.Decimal
	for _, ask := range o.Asks {
		if !IsArbGreaterThanThreshold(ask.Price, o.Bids[bidIndex].Price, limits.FeeRateBps, limits.ProfitThresholdBps) {
			// if there is no arb at this price, there is definitely no arb at a worse price
			break
		}
		askSliceQuantity := decimal.Min(remainingAvailableQuote.Div(ask.Price), ask.Quantity)
		highestBuyPrice = ask.Price
		cumulativeAskQuantitySold := decimal.Zero
		for {
			bidSliceQuantity = decimal.Min(o.Bids[bidIndex].Quantity, askSliceQuantity) // this is the amount we can purchase from this ask and offload on this bid
			o.Bids[bidIndex].Quantity = o.Bids[bidIndex].Quantity.Sub(bidSliceQuantity)
			lowestSellPrice = o.Bids[bidIndex].Price
			// update cumulative bought
			cumulativeQuantityBought = cumulativeQuantityBought.Add(bidSliceQuantity)
			cumulativeBuyCost = cumulativeBuyCost.Add(bidSliceQuantity.Mul(ask.Price))
			cumulativeBuyCostWFees = cumulativeBuyCostWFees.Add(bidSliceQuantity.Mul(ask.Price).Mul(tc.One.Add(limits.FeeRateBps.Div(tc.OneE5))))
			// update cumulative sold
			cumulativeQuantitySold = cumulativeQuantitySold.Add(bidSliceQuantity)
			cumulativeAskQuantitySold = cumulativeAskQuantitySold.Add(bidSliceQuantity)
			cumulativeProceedsFromSale = cumulativeProceedsFromSale.Add(bidSliceQuantity.Mul(o.Bids[bidIndex].Price))
			cumulativeProceedsFromSaleWFees = cumulativeProceedsFromSaleWFees.Add(bidSliceQuantity.Mul(o.Bids[bidIndex].Price).Mul(tc.One.Add(limits.FeeRateBps.Div(tc.OneE5))))
			// decrement quantity
			remainingAvailableQuote = remainingAvailableQuote.Sub(bidSliceQuantity.Mul(ask.Price))
			// incremement the bids if we've sold through one
			if o.Bids[bidIndex].Quantity.Equal(decimal.Zero) {
				bidIndex++
			} else if o.Bids[bidIndex].Quantity.LessThan(decimal.Zero) {
				fmt.Println("BAD ERROR SHOULD NEVER HAPPEN")
			}
			// break if we've sold everything we bought, or if there is not an arb further into the book
			if cumulativeAskQuantitySold.GreaterThanOrEqual(askSliceQuantity) || !IsArbGreaterThanThreshold(ask.Price, o.Bids[bidIndex].Price, limits.FeeRateBps, limits.ProfitThresholdBps) {
				break
			}
		}
		if remainingAvailableQuote.Equal(decimal.Zero) {
			break
		}
	}
	if cumulativeQuantityBought.LessThanOrEqual(decimal.Zero) {
		err = errNoArb
		return
	}
	buyVWAP := cumulativeBuyCostWFees.Div(cumulativeQuantityBought)
	sellVWAP := cumulativeProceedsFromSaleWFees.Div(cumulativeQuantitySold)
	maxQAtLimitBuy := decimal.Min(cumulativeQuantityBought, decimal.Min(limits.MaxOrderAmount, availableQuoteBalance).Div(highestBuyPrice))
	quantityToBuy := maxQAtLimitBuy.Truncate(5)
	if quantityToBuy.LessThanOrEqual(decimal.Zero) {
		err = errNoArb
		return
	}
	profit := sellVWAP.Sub(buyVWAP).Mul(quantityToBuy)
	profitBps := profit.Div(buyVWAP.Mul(quantityToBuy)).Mul(tc.OneE5)
	buyLimit := highestBuyPrice.Truncate(8)
	sellLimit := lowestSellPrice.Truncate(8)
	// fmt.Println(inOb.Pair.String(), "arb: ", profitBps)
	if profit.LessThanOrEqual(decimal.Zero) || profitBps.LessThan(limits.ProfitThresholdBps) ||
		quantityToBuy.LessThan(limits.MinOrderQuantity) || quantityToBuy.GreaterThanOrEqual(limits.MaxOrderQuantity) ||
		quantityToBuy.Mul(buyLimit).LessThan(limits.MinOrderAmount) {
		err = errNoArb
		return
	}
	arb = arbStrat{
		Pair:           o.Pair,
		BuyPrice:       buyVWAP,
		SellPrice:      sellVWAP,
		BuyLimitPrice:  buyLimit,
		SellLimitPrice: sellLimit,
		Quantity:       quantityToBuy,
		ProfitGoal:     profit,
		ProfitGoalBps:  profitBps,
	}
	return
}

func IsArbGreaterThanThreshold(priceBuy, priceSell, feeRateBps, profitThresholdBps decimal.Decimal) bool {
	adjustedBuyPrice := priceBuy.Mul(tc.One.Add(feeRateBps.Div(tc.OneE5)))
	adjustedSellPrice := priceSell.Mul(tc.One.Add(feeRateBps.Div(tc.OneE5)))
	return adjustedSellPrice.Sub(adjustedBuyPrice).Div(adjustedBuyPrice).Mul(tc.OneE5).GreaterThanOrEqual(profitThresholdBps)
}
