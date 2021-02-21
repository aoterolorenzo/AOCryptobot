package interfaces

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/models"
)

type ExchangeService interface {
	SetPair(pair string)
	ConfigureClient()
	GetTotalBalance(asset string) (float64, error)
	GetAvailableBalance(asset string) (float64, error)
	GetLockedBalance(asset string) (float64, error)
	MakeOrder(quantity float64, rate float64, orderSide techan.OrderSide) (models.Order, error)
	MakeOCOOrder(quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
		orderSide techan.OrderSide) (models.OCOOrder, error)
	GetOrder(orderId int64) (models.Order, error)
	CancelOrder(orderId int64) error
	GetOrderStatus(orderId int64) (models.OrderStatusType, error)
	DepthMonitor(marketSnapshotsRecord *[]models.MarketDepth)
	TimeSeriesMonitor(interval string, timeSeries *techan.TimeSeries)
}
