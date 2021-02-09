package common

import (
	"../../model"
	"github.com/adshao/go-binance/v2"
)

type OrderBookService struct {
	OrderBook model.OrderBook
}

func (ob *OrderBookService) Init() {
	ob.OrderBook = model.OrderBook{}
}

func (ob *OrderBookService) AddFilledOrder(order binance.Order) {
	ob.OrderBook.FilledOrders = append(ob.OrderBook.FilledOrders, order)
}

func (ob *OrderBookService) AddOpenOrder(order binance.Order) {
	ob.OrderBook.OpenOrders = append(ob.OrderBook.OpenOrders, order)
}

func (ob *OrderBookService) AddCanceledOrder(order binance.Order) {
	ob.OrderBook.CanceledOrders = append(ob.OrderBook.CanceledOrders, order)
}

func (ob *OrderBookService) RemoveOpenOrder(order binance.Order) {
	for i, lOrder := range ob.OrderBook.OpenOrders {
		if order == lOrder {
			(&ob.OrderBook).OpenOrders = append(((&ob.OrderBook).OpenOrders)[:i], ((&ob.OrderBook).OpenOrders)[i+1:]...)
			break
		}
	}
}

func (ob *OrderBookService) OpenSellOrdersCount() int {
	count := 0
	if ob.OrderBook.OpenOrders != nil {
		for _, book := range ob.OrderBook.OpenOrders {
			if book.Side == "SELL" {
				count += 1
			}
		}
	}
	return count
}

func (ob *OrderBookService) OpenBuyOrdersCount() int {
	count := 0
	if ob.OrderBook.OpenOrders != nil {
		for _, book := range ob.OrderBook.OpenOrders {
			if book.Side == "BUY" {
				count++
			}
		}
	}
	return count
}

func (ob *OrderBookService) OpenOrdersCount() int {
	if ob.OrderBook.OpenOrders != nil {
		return len(ob.OrderBook.OpenOrders)
	}
	return 0
}
