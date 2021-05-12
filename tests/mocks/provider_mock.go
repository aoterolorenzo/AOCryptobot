package mocks

import (
	"github.com/adshao/go-binance/v2"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gitlab.com/aoterocom/AOCryptobot/tests/utils"
	"log"
)

type ProviderMock struct {
	timeSeries *techan.TimeSeries
}

func NewPaperService() *ProviderMock {
	return &ProviderMock{}
}

func init() {
}

func (paperService *ProviderMock) GetTotalBalance(asset string) (float64, error) {
	return 1000.0, nil
}

func (paperService *ProviderMock) GetAvailableBalance(asset string) (float64, error) {
	return 1000.0, nil
}

func (paperService *ProviderMock) GetLockedBalance(asset string) (float64, error) {
	return 1000.0, nil
}

func (paperService *ProviderMock) MakeOrder(pair string, quantity float64, rate float64,
	orderType models.OrderType, orderSide models.OrderSide) (models.Order, error) {
	order := models.Order{}
	return order, nil
}

func (paperService *ProviderMock) MakeOCOOrder(pair string, quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
	orderSide models.OrderSide) (models.OCOOrder, error) {
	return models.OCOOrder{}, nil
}

func (paperService *ProviderMock) GetOrder(order models.Order) (models.Order, error) {
	return models.NewEmptyOrder(), nil
}

func (paperService *ProviderMock) CancelOrder(order models.Order) error {
	return nil
}

func (paperService *ProviderMock) GetOrderStatus(order models.Order) (models.OrderStatusType, error) {
	return models.OrderStatusTypeFilled, nil
}

func (paperService *ProviderMock) DepthMonitor(pair string, marketSnapshotsRecord *[]models.MarketDepth) {
}

func (paperService *ProviderMock) TimeSeriesMonitor(pair string, interval string, timeSeries *techan.TimeSeries, active *bool) {
}

func (paperService *ProviderMock) GetSeries(pair string, interval string, limit int) (techan.TimeSeries, error) {
	var series techan.TimeSeries
	if err := utils.Load("../resources/ETHEUR.json", &series); err != nil {
		log.Fatalln(err)
	}
	return series, nil
}

func (paperService *ProviderMock) GetMarkets(coin string) []string {
	var pairList []string
	return pairList
}

func (paperService *ProviderMock) GetPairInfo(pair string) *models.PairInfo {
	return nil
}

func (paperService *ProviderMock) wsKlineHandler(event *binance.WsKlineEvent) {
}

func (paperService *ProviderMock) errHandler(err error) {
	helpers.Logger.Errorln(err)
}
