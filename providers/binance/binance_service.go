package binance

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"os"
	"reflect"
	"strconv"
	"time"
)

var logger = helpers.Logger{}

type BinanceService struct {
	binanceClient         *binance.Client
	timeSeries            *techan.TimeSeries
	marketSnapshotsRecord *[]models.MarketDepth
	apiKey                string
	apiSecret             string
	pair                  string
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/providers/binance/conf.env")
	if err != nil {
		logger.Fatalln("Error loading go.env file", err)
	}
}

func (binanceService *BinanceService) SetPair(pair string) {
	binanceService.pair = pair
}

func (binanceService *BinanceService) ConfigureClient() {
	binanceService.apiKey = os.Getenv("apiKey")
	binanceService.apiSecret = os.Getenv("apiSecret")
	binanceService.binanceClient = binance.NewClient(binanceService.apiKey, binanceService.apiSecret)
}

func (binanceService *BinanceService) GetTotalBalance(asset string) (float64, error) {
	res, err := binanceService.binanceClient.NewGetAccountService().Do(context.Background())
	if err != nil {
		return 0, nil
	}
	for _, v := range res.Balances {
		if v.Asset == asset {

			free, err := strconv.ParseFloat(v.Free, 64)
			if err != nil {
				return 0, nil
			}

			locked, err := strconv.ParseFloat(v.Locked, 64)
			if err != nil {
				return 0, nil
			}

			return free + locked, nil
		}
	}

	return -1.0, fmt.Errorf("error: unknown error getting through the balances")
}

func (binanceService *BinanceService) GetAvailableBalance(asset string) (float64, error) {
	res, err := binanceService.binanceClient.NewGetAccountService().Do(context.Background())
	if err != nil {
		return 0, err
	}
	for _, v := range res.Balances {
		if v.Asset == asset {

			free, err := strconv.ParseFloat(v.Free, 64)
			if err != nil {
				return 0, err
			}

			return free, nil
		}
	}

	return -1.0, fmt.Errorf("error: unknown error getting through the balances")
}

func (binanceService *BinanceService) GetLockedBalance(asset string) (float64, error) {
	res, err := binanceService.binanceClient.NewGetAccountService().Do(context.Background())
	if err != nil {
		return 0, nil
	}
	for _, v := range res.Balances {
		if v.Asset == asset {

			locked, err := strconv.ParseFloat(v.Locked, 64)
			if err != nil {
				return 0, nil
			}

			return locked, nil
		}
	}

	return -1.0, fmt.Errorf("error: unknown error getting through the balances")
}

func (binanceService *BinanceService) MakeOrder(quantity float64, rate float64,
	orderSide techan.OrderSide) (models.Order, error) {

	if orderSide == techan.BUY {
		quantity = quantity / rate
	}

	//Convert techan orderSide to binance SideType
	var sideType binance.SideType
	if orderSide == techan.BUY {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	order, err := binanceService.binanceClient.NewCreateOrderService().Symbol(binanceService.pair).
		Side(sideType).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(fmt.Sprintf("%.5f", quantity)).
		Price(fmt.Sprintf("%.2f", rate)).Do(context.Background())

	if err != nil {
		return models.Order{}, err
	}

	return binanceService.orderResponseToOrder(*order), nil
}

func (binanceService *BinanceService) MakeOCOOrder(quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
	orderSide techan.OrderSide) (models.OCOOrder, error) {

	if orderSide == techan.BUY {
		quantity = quantity / rate
	}

	//Convert techan orderSide to binance SideType
	var sideType binance.SideType
	if orderSide == techan.BUY {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	order, err := binanceService.binanceClient.NewCreateOCOService().Symbol(binanceService.pair).Side(sideType).
		Price(fmt.Sprintf("%.2f", rate)).
		StopPrice(fmt.Sprintf("%.2f", stopPrice)).
		StopLimitPrice(fmt.Sprintf("%.2f", stopLimitPrice)).
		Quantity(fmt.Sprintf("%.5f", quantity)).
		StopLimitTimeInForce("GTC").Do(context.Background())

	if err != nil {
		return models.OCOOrder{}, err
	}

	return binanceService.ocoOrderResponseToOCOOrder(*order), nil
}

func (binanceService *BinanceService) GetOrder(orderId int64) (models.Order, error) {
	order, err := binanceService.binanceClient.NewGetOrderService().Symbol(binanceService.pair).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return models.Order{}, err
	}

	return binanceService.orderToModelsOrder(*order), nil
}

func (binanceService *BinanceService) CancelOrder(orderId int64) error {
	_, err := binanceService.binanceClient.NewCancelOrderService().Symbol(binanceService.pair).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (binanceService *BinanceService) GetOrderStatus(orderId int64) (models.OrderStatusType, error) {
	order, err := binanceService.GetOrder(orderId)
	if err != nil {
		return models.OrderStatusTypeNew, err
	}

	return order.Status, err
}

func (binanceService *BinanceService) DepthMonitor(marketSnapshotsRecord *[]models.MarketDepth) {
	binanceService.marketSnapshotsRecord = marketSnapshotsRecord
	doneC, _, err := binance.WsDepthServe(binanceService.pair, binanceService.wsDepthHandler, binanceService.errHandler)
	if err != nil {
		logger.Errorln(err)
		return
	}
	<-doneC
}

func (binanceService *BinanceService) TimeSeriesMonitor(interval string, timeSeries *techan.TimeSeries) {
	binanceService.timeSeries = timeSeries

	klines, err := binanceService.binanceClient.NewKlinesService().Symbol(binanceService.pair).
		Interval(interval).Do(context.Background())
	if err != nil {
		logger.Fatalln("error getting klines: " + err.Error())
	}

	for _, k := range klines {
		period := techan.NewTimePeriod(time.Unix(k.OpenTime/1000, 0), time.Minute*15)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(k.Open)
		candle.ClosePrice = big.NewFromString(k.Close)
		candle.MaxPrice = big.NewFromString(k.High)
		candle.MinPrice = big.NewFromString(k.Low)
		candle.TradeCount = uint(k.TradeNum)
		candle.Volume = big.NewFromString(k.Volume)
		binanceService.timeSeries.AddCandle(candle)
	}

	doneC, _, err := binance.WsKlineServe(binanceService.pair, interval, binanceService.wsKlineHandler, binanceService.errHandler)
	if err != nil {
		logger.Errorln(err)
		return
	}
	<-doneC
}

func (binanceService *BinanceService) orderResponseToOrder(o binance.CreateOrderResponse) models.Order {
	return models.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.OrderID,
		ClientOrderID:            o.ClientOrderID,
		Price:                    o.Price,
		OrigQuantity:             o.OrigQuantity,
		ExecutedQuantity:         o.ExecutedQuantity,
		CummulativeQuoteQuantity: o.CummulativeQuoteQuantity,
		Status:                   models.OrderStatusType(o.Status),
		Type:                     models.OrderType(o.Type),
		Side:                     models.SideType(o.Side),
		StopPrice:                "",
		IcebergQuantity:          "",
		Time:                     o.TransactTime,
		UpdateTime:               0,
		IsWorking:                false,
	}
}

func (binanceService *BinanceService) orderToModelsOrder(o binance.Order) models.Order {
	return models.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.OrderID,
		ClientOrderID:            o.ClientOrderID,
		Price:                    o.Price,
		OrigQuantity:             o.OrigQuantity,
		ExecutedQuantity:         o.ExecutedQuantity,
		CummulativeQuoteQuantity: o.CummulativeQuoteQuantity,
		Status:                   models.OrderStatusType(o.Status),
		Type:                     models.OrderType(o.Type),
		Side:                     models.SideType(o.Side),
		StopPrice:                "",
		IcebergQuantity:          "",
		UpdateTime:               0,
		IsWorking:                false,
	}
}

func (binanceService *BinanceService) ocoOrderToModelsOrder(o binance.OCOOrder) models.Order {
	return models.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.OrderID,
		ClientOrderID:            o.ClientOrderID,
		Price:                    "",
		OrigQuantity:             "",
		ExecutedQuantity:         "",
		CummulativeQuoteQuantity: "",
		Status:                   "",
		Type:                     "",
		Side:                     "",
		StopPrice:                "",
		IcebergQuantity:          "",
		Time:                     0,
		UpdateTime:               0,
		IsWorking:                false,
		IsIsolated:               false,
	}
}

func (binanceService *BinanceService) orderListToModelsOrderList(ol []*binance.Order) []models.Order {
	var modelsOrderList []models.Order
	for _, binOrder := range ol {
		modelsOrderList = append(modelsOrderList, binanceService.orderToModelsOrder(*binOrder))
	}
	return modelsOrderList
}

func (binanceService *BinanceService) ocoOrderResponseToOCOOrder(o binance.CreateOCOResponse) models.OCOOrder {
	OCOOrder := models.OCOOrder{
		OrderListID:       o.OrderListID,
		ContingencyType:   o.ContingencyType,
		ListStatusType:    o.ListStatusType,
		ListOrderStatus:   o.ListOrderStatus,
		ListClientOrderID: o.ListClientOrderID,
		TransactionTime:   o.TransactionTime,
		Symbol:            o.Symbol,
	}

	for _, binOrder := range o.Orders {
		OCOOrder.Orders = append(OCOOrder.Orders, binanceService.ocoOrderToModelsOrder(*binOrder))
	}

	return OCOOrder
}

func (binanceService *BinanceService) wsKlineHandler(event *binance.WsKlineEvent) {
	lastCandle := binanceService.timeSeries.Candles[len(binanceService.timeSeries.Candles)-1]

	period := techan.NewTimePeriod(time.Unix(event.Kline.StartTime/1000, 0), time.Minute*15)
	candle := techan.NewCandle(period)
	candle.OpenPrice = big.NewFromString(event.Kline.Open)
	candle.ClosePrice = big.NewFromString(event.Kline.Close)
	candle.MaxPrice = big.NewFromString(event.Kline.High)
	candle.MinPrice = big.NewFromString(event.Kline.Low)
	candle.TradeCount = uint(event.Kline.TradeNum)
	candle.Volume = big.NewFromString(event.Kline.Volume)

	if lastCandle.Period != techan.NewTimePeriod(time.Unix(event.Kline.StartTime/1000, 0), time.Minute*15) {
		binanceService.timeSeries.AddCandle(candle)
	} else if lastCandle.Period == techan.NewTimePeriod(time.Unix(event.Kline.StartTime/1000, 0), time.Minute*15) {
		binanceService.timeSeries.Candles[len(binanceService.timeSeries.Candles)-1] = candle
	}
}

func (binanceService *BinanceService) wsDepthHandler(event *binance.WsDepthEvent) {
	marketSnapshot := models.MarketDepth{}
	err := marketSnapshot.Set(event)
	if err != nil {
		binanceService.errHandler(err)
		return
	}

	// Grab the record only if there are 2 bid and ask groups
	if marketSnapshot.WsDepthEvent != nil && marketSnapshot.WsDepthEvent.Bids != nil && marketSnapshot.WsDepthEvent.Asks != nil &&
		len(marketSnapshot.WsDepthEvent.Bids) > 2 && len(marketSnapshot.WsDepthEvent.Asks) > 2 {

		//Append to MarkerSnapshotsRecord
		reverseAny(*binanceService.marketSnapshotsRecord)
		*binanceService.marketSnapshotsRecord = append(*binanceService.marketSnapshotsRecord, marketSnapshot)
		reverseAny(*binanceService.marketSnapshotsRecord)

	}
}

func (binanceService *BinanceService) errHandler(err error) {
	logger.Errorln(err)
}

func reverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
