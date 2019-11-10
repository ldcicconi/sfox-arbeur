package trader

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
