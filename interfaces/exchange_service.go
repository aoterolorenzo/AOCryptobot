package interfaces

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/models"
)

type ExchangeService interface {
	GetTotalBalance(asset string) (float64, error)
	GetAvailableBalance(asset string) (float64, error)
	GetLockedBalance(asset string) (float64, error)
	MakeOrder(pair string, quantity float64, rate float64,
		orderType models.OrderType, orderSide models.OrderSide) (models.Order, error)
	MakeOCOOrder(pair string, quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
		orderSide models.OrderSide) (models.OCOOrder, error)
	GetOrder(order models.Order) (models.Order, error)
	CancelOrder(order models.Order) error
	GetOrderStatus(order models.Order) (models.OrderStatusType, error)
	DepthMonitor(pair string, marketSnapshotsRecord *[]models.MarketDepth)
	TimeSeriesMonitor(pair string, interval string, timeSeries *techan.TimeSeries, active *bool)
	GetSeries(pair string, interval string, limit int) (techan.TimeSeries, error)
	GetMarkets(coin string, whitelist []string, blacklist []string) []string
	GetPairInfo(pair string) *models.PairInfo
}
