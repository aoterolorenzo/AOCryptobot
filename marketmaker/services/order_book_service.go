package services

import (
	"github.com/adshao/go-binance/v2"
	"gitlab.com/aoterocom/AOCryptobot/marketmaker/models"
	"sync"
)

type OrderBookService struct {
	OrderBook models.OrderBook
	mutex     *sync.Mutex
}

func (ob *OrderBookService) Init() {
	ob.OrderBook = models.OrderBook{}
}

func (ob *OrderBookService) SetMutex(mutex *sync.Mutex) {
	ob.mutex = mutex
}

func (ob *OrderBookService) AddFilledOrder(order binance.Order) {
	ob.mutex.Lock()
	ob.OrderBook.FilledOrders = append(ob.OrderBook.FilledOrders, order)
	ob.mutex.Unlock()
}

func (ob *OrderBookService) AddOpenOrder(order binance.Order) {
	ob.mutex.Lock()
	ob.OrderBook.OpenOrders = append(ob.OrderBook.OpenOrders, order)
	ob.mutex.Unlock()
}

func (ob *OrderBookService) AddCanceledOrder(order binance.Order) {
	ob.mutex.Lock()
	ob.OrderBook.CanceledOrders = append(ob.OrderBook.CanceledOrders, order)
	ob.mutex.Unlock()
}

func (ob *OrderBookService) RemoveOpenOrder(order binance.Order) {
	ob.mutex.Lock()
	for i, lOrder := range ob.OrderBook.OpenOrders {
		if order == lOrder {
			(&ob.OrderBook).OpenOrders = append(((&ob.OrderBook).OpenOrders)[:i], ((&ob.OrderBook).OpenOrders)[i+1:]...)
			break
		}
	}
	ob.mutex.Unlock()
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