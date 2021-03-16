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
	"strings"
	"time"
)

type BinanceService struct {
	binanceClient         *binance.Client
	timeSeries            *techan.TimeSeries
	marketSnapshotsRecord *[]models.MarketDepth
	apiKey                string
	apiSecret             string
}

func NewBinanceService() BinanceService {
	binanceService := BinanceService{}
	binanceService.apiKey = os.Getenv("apiKey")
	binanceService.apiSecret = os.Getenv("apiSecret")
	binanceService.binanceClient = binance.NewClient(binanceService.apiKey, binanceService.apiSecret)
	return binanceService
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/providers/binance/conf.env")
	if err != nil {
		helpers.Logger.Fatalln("Error loading go.env file", err)
	}
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

func (binanceService *BinanceService) MakeOrder(pair string, quantity float64, rate float64,
	orderType models.OrderType, orderSide models.OrderSide) (models.Order, error) {

	if orderSide == models.BUY {
		quantity = quantity / rate
	}

	//Convert techan orderSide to binance SideType
	var sideType binance.SideType
	if orderSide == models.BUY {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	//Convert models orderType to binance orderType
	binanceOrderType := binance.OrderType(orderType)

	preparedOrder := binanceService.binanceClient.NewCreateOrderService().Symbol(pair).
		Side(sideType).Type(binanceOrderType).Quantity(fmt.Sprintf("%.5f", quantity))

	if binanceOrderType == binance.OrderTypeLimit {
		order, err := preparedOrder.
			TimeInForce(binance.TimeInForceTypeGTC).Price(fmt.Sprintf("%.2f", rate)).Do(context.Background())
		if err != nil {
			return models.NewEmptyOrder(), err
		}
		return binanceService.orderResponseToOrder(*order), nil
	} else {
		order, err := preparedOrder.Do(context.Background())
		if err != nil {
			return models.NewEmptyOrder(), err
		}
		return binanceService.orderResponseToOrder(*order), nil
	}

}

func (binanceService *BinanceService) MakeOCOOrder(pair string, quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
	orderSide models.OrderSide) (models.OCOOrder, error) {

	if orderSide == models.BUY {
		quantity = quantity / rate
	}

	//Convert orderSide to binance SideType
	var sideType binance.SideType
	if orderSide == models.BUY {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	order, err := binanceService.binanceClient.NewCreateOCOService().Symbol(pair).Side(sideType).
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

func (binanceService *BinanceService) GetOrder(order models.Order) (models.Order, error) {
	responseOrder, err := binanceService.binanceClient.NewGetOrderService().Symbol(order.Symbol).
		OrderID(order.OrderID).Do(context.Background())
	if err != nil {
		return models.NewEmptyOrder(), err
	}

	return binanceService.orderToModelsOrder(*responseOrder), nil
}

func (binanceService *BinanceService) CancelOrder(order models.Order) error {
	_, err := binanceService.binanceClient.NewCancelOrderService().Symbol(order.Symbol).
		OrderID(order.OrderID).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (binanceService *BinanceService) GetOrderStatus(order models.Order) (models.OrderStatusType, error) {
	responseOrder, err := binanceService.GetOrder(order)
	if err != nil {
		return models.OrderStatusTypeNew, err
	}

	return responseOrder.Status, err
}

func (binanceService *BinanceService) DepthMonitor(pair string, marketSnapshotsRecord *[]models.MarketDepth) {
	binanceService.marketSnapshotsRecord = marketSnapshotsRecord
	doneC, _, err := binance.WsDepthServe(pair, binanceService.wsDepthHandler, binanceService.errHandler)
	if err != nil {
		helpers.Logger.Errorln(err)
		return
	}
	<-doneC
}

func (binanceService *BinanceService) TimeSeriesMonitor(pair, interval string, timeSeries *techan.TimeSeries, active *bool) {
	binanceService.timeSeries = timeSeries

	klines, err := binanceService.binanceClient.NewKlinesService().Symbol(pair).
		Interval(interval).Do(context.Background())
	if err != nil {
		helpers.Logger.Fatalln("error getting klines: " + err.Error())
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

	doneC, done, err := binance.WsKlineServe(pair, interval, binanceService.wsKlineHandler, binanceService.errHandler)
	if err != nil {
		helpers.Logger.Errorln(err)
		return
	}

	go func() {
		for *active {
			time.Sleep(1 * time.Second)
		}
		done <- struct{}{}
	}()

	<-doneC
}

func (binanceService *BinanceService) GetSeries(pair string, interval string, limit int) (techan.TimeSeries, error) {
	if limit == 0 {
		limit = 1000
	}
	timeSeries := techan.TimeSeries{}
	klines, err := binanceService.binanceClient.NewKlinesService().Symbol(pair).
		Interval(interval).Limit(limit).Do(context.Background())
	if err != nil {
		return timeSeries, err
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
		timeSeries.AddCandle(candle)
	}

	return timeSeries, nil
}

func (binanceService *BinanceService) GetMarkets(coin string) []string {
	var pairList []string
	info, _ := binanceService.binanceClient.NewExchangeInfoService().Do(context.Background())
	for _, symbol := range info.Symbols {
		if strings.Contains(symbol.Symbol, coin) {
			pairList = append(pairList, symbol.Symbol)
		}
	}
	return pairList
}

func (binanceService *BinanceService) orderResponseToOrder(o binance.CreateOrderResponse) models.Order {
	return models.NewOrder(o.Symbol, o.OrderID, o.ClientOrderID, o.Price, o.OrigQuantity, o.ExecutedQuantity,
		o.CummulativeQuoteQuantity, models.OrderStatusType(o.Status), models.OrderType(o.Type), models.SideType(o.Side),
		o.TransactTime, 0, false, false)
}

func (binanceService *BinanceService) orderToModelsOrder(o binance.Order) models.Order {
	return models.NewOrder(o.Symbol, o.OrderID, o.ClientOrderID, o.Price, o.OrigQuantity, o.ExecutedQuantity,
		o.CummulativeQuoteQuantity, models.OrderStatusType(o.Status), models.OrderType(o.Type), models.SideType(o.Side),
		0, 0, false, false)
}

func (binanceService *BinanceService) ocoOrderToModelsOrder(o binance.OCOOrder) models.Order {
	return models.NewOrder(o.Symbol, o.OrderID, o.ClientOrderID, "", "", "",
		"", "", "", "",
		0, 0, false, false)
}

func (binanceService *BinanceService) orderListToModelsOrderList(ol []*binance.Order) []models.Order {
	var modelsOrderList []models.Order
	for _, binOrder := range ol {
		modelsOrderList = append(modelsOrderList, binanceService.orderToModelsOrder(*binOrder))
	}
	return modelsOrderList
}

func (binanceService *BinanceService) ocoOrderResponseToOCOOrder(o binance.CreateOCOResponse) models.OCOOrder {
	OCOOrder := models.NewOCOOrder(o.OrderListID, o.ContingencyType, o.ListStatusType, o.ListOrderStatus,
		o.ListClientOrderID, o.TransactionTime, o.Symbol)

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
	marketSnapshot := models.NewMarketDepth()
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
	helpers.Logger.Errorln(err)
}

func reverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
