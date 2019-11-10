package execution

type OrderStatusCode int

type OrderStatusEnvelope struct {
	StatusCode OrderStatusCode
	Details    interface{}
}

var (
	OSTATUS_SENT         = OrderStatusCode(100)
	OSTATUS_SUCCESS      = OrderStatusCode(200)
	OSTATUS_PARTIAL_FILL = OrderStatusCode(300)
	OSTATUS_DONE         = OrderStatusCode(400)
	OSTATUS_CANCELED     = OrderStatusCode(500)
)
