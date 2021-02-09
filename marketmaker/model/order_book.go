package model

import (
	"github.com/adshao/go-binance/v2"
)

type OrderBook struct {
	FilledOrders             []binance.Order
	OpenOrders               []binance.Order
	CanceledOrders           []binance.Order
	InitialPairPrice         float64
	InitialCoin1FreeAmount   float64
	InitialCoin1LockedAmount float64
	InitialCoin2FreeAmount   float64
	InitialCoin2LockedAmount float64
}
