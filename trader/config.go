package trader

import (
	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

type TraderConfig struct {
	Pair                tc.Pair
	MaxPositionQuantity decimal.Decimal // in quote currency of the pair
	ProfitThresholdBps  decimal.Decimal
	FeeRateBps          decimal.Decimal
}
