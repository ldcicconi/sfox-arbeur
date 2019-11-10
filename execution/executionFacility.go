package execution

import "github.com/ldcicconi/sfox-arbeur/trader"

type Broker interface {
	StartExecutionDeal() *ExecutionDeal
}

type ExecutionDeal struct {
	OrderChannel    chan trader.TraderOrder  // This is the channel the trader should send new orders to
	OrderStatusChan chan OrderStatusEnvelope // This is the channel the broker sends order status back to the traders
}

func NewBroker(exchangeName string, apiKey string) Broker {
	switch exchangeName {
	case "sfox":
		return NewSFOXBroker(apiKey)
	default:
		return nil
	}
}
