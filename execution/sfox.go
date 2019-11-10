package execution

import "github.com/ldcicconi/sfox-arbeur/trader"

type SFOXBroker struct {
	newOrderChannel    chan trader.TraderOrder
	orderStatusChannel chan OrderStatusEnvelope
	clientPool         SFOXAPIClientPool
}

func NewSFOXBroker(APIKey string) *SFOXBroker {
	return &SFOXBroker{}
}

func (sfox *SFOXBroker) StartExecutionDeal() *ExecutionDeal {
	go sfox.monitorOrderChannel()
	return &ExecutionDeal{
		sfox.newOrderChannel,
		sfox.orderStatusChannel,
	}
}

func (sfox *SFOXBroker) monitorOrderChannel() {
	go func() {
		for o := range sfox.newOrderChannel {
			sfox.handleOrder(o)
		}
	}()
}

func (sfox *SFOXBroker) handleOrder(o trader.TraderOrder) {
	client := sfox.clientPool.GetAPIClient()
	client.NewOrder(o.Quantity, o.LimitPrice, o.AlgoID, o.Pair.String(), string(o.Side))
}
