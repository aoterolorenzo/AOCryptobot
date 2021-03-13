package models

type OrderBook struct {
	FilledOrders             []Order
	OpenOrders               []Order
	CanceledOrders           []Order
	InitialPairPrice         float64
	InitialCoin1FreeAmount   float64
	InitialCoin1LockedAmount float64
	InitialCoin2FreeAmount   float64
	InitialCoin2LockedAmount float64
}

func NewOrderBook() OrderBook {
	return OrderBook{}
}
