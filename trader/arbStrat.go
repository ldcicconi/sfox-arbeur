package trader

import (
	"fmt"

	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

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
}

func FindArb(o tc.SFOXOrderbook, feeRateBps, profitMinBps, maxPositionSize decimal.Decimal) *arbStrat {
	arb := o.Arb()
	if arb.LessThanOrEqual(decimal.Zero) {
		return nil
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
	remainingQuantity := maxPositionSize

	var bidSliceQuantity decimal.Decimal
	for _, ask := range o.Asks {
		if !IsArbGreaterThanThreshold(ask.Price, o.Bids[bidIndex].Price, feeRateBps, profitMinBps) {
			// if there is no arb at this price, there is definitely no arb at a worse price
			break
		}
		sliceQuantity := decimal.Min(remainingQuantity, ask.Quantity)
		highestBuyPrice = ask.Price
		for {
			bidSliceQuantity = decimal.Min(o.Bids[bidIndex].Quantity, sliceQuantity)
			o.Bids[bidIndex].Quantity = o.Bids[bidIndex].Quantity.Sub(bidSliceQuantity)
			if !bidSliceQuantity.Equal(decimal.Zero) {
				lowestSellPrice = o.Bids[bidIndex].Price
			}
			// update cumulative bought
			cumulativeQuantityBought = cumulativeQuantityBought.Add(bidSliceQuantity)
			cumulativeBuyCost = cumulativeBuyCost.Add(bidSliceQuantity.Mul(ask.Price))
			cumulativeBuyCostWFees = cumulativeBuyCostWFees.Add(bidSliceQuantity.Mul(ask.Price).Mul(tc.One.Add(feeRateBps.Mul(tc.OneE5))))
			// update cumulative sold
			cumulativeQuantitySold = cumulativeQuantitySold.Add(bidSliceQuantity)
			cumulativeProceedsFromSale = cumulativeProceedsFromSale.Add(bidSliceQuantity.Mul(o.Bids[bidIndex].Price))
			cumulativeProceedsFromSaleWFees = cumulativeProceedsFromSaleWFees.Add(bidSliceQuantity.Mul(o.Bids[bidIndex].Price).Mul(tc.One.Sub(feeRateBps.Mul(tc.OneE5))))
			// decrement quantity
			remainingQuantity = remainingQuantity.Sub(bidSliceQuantity)
			// incremement the bids if we've sold through one
			if o.Bids[bidIndex].Quantity.Equal(decimal.Zero) {
				bidIndex++
			} else if o.Bids[bidIndex].Quantity.LessThan(decimal.Zero) {
				fmt.Println("BAD ERROR SHOULD NEVER HAPPEN")
			}
			// break if we've sold everything we bought, or if there is not an arb further into the book
			if cumulativeQuantitySold.GreaterThanOrEqual(sliceQuantity) || !IsArbGreaterThanThreshold(ask.Price, o.Bids[bidIndex].Price, feeRateBps, profitMinBps) {
				break
			}
		}
		if remainingQuantity.Equal(decimal.Zero) {
			break
		}
	}
	if cumulativeQuantityBought.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	buyVWAP := cumulativeBuyCostWFees.Div(cumulativeQuantityBought)
	sellVWAP := cumulativeProceedsFromSaleWFees.Div(cumulativeQuantitySold)
	return &arbStrat{
		Pair:           o.Pair,
		BuyPrice:       buyVWAP,
		SellPrice:      sellVWAP,
		BuyLimitPrice:  highestBuyPrice,
		SellLimitPrice: lowestSellPrice,
		Quantity:       cumulativeQuantityBought,
		ProfitGoalBps:  sellVWAP.Sub(buyVWAP).Div(buyVWAP).Mul(tc.OneE5),
	}
}

func IsArbGreaterThanThreshold(priceBuy, priceSell, feeRateBps, profitThresholdBps decimal.Decimal) bool {
	adjustedBuyPrice := priceBuy.Mul(tc.One.Add(feeRateBps.Div(tc.OneE5)))
	adjustedSellPrice := priceSell.Mul(tc.One.Add(feeRateBps.Div(tc.OneE5)))
	return adjustedSellPrice.Sub(adjustedBuyPrice).Div(adjustedBuyPrice).Mul(tc.OneE5).GreaterThanOrEqual(profitThresholdBps)
}
