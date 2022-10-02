package binance

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/database"
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
	dBService             *database.DBService
	timeSeries            *techan.TimeSeries
	marketSnapshotsRecord *[]models.MarketDepth
	apiKey                string
	apiSecret             string
	pair                  string
	interval              string
	active                *bool
}

func NewBinanceService() *BinanceService {
	binanceService := BinanceService{}
	binanceService.apiKey = os.Getenv("binanceAPIKey")
	binanceService.apiSecret = os.Getenv("binanceAPISecret")
	binanceService.binanceClient = binance.NewClient(binanceService.apiKey, binanceService.apiSecret)
	return &binanceService
}

func NewBinanceDBService(databaseService *database.DBService) *BinanceService {
	binanceService := BinanceService{
		dBService: databaseService,
	}
	binanceService.apiKey = os.Getenv("binanceAPIKey")
	binanceService.apiSecret = os.Getenv("binanceAPISecret")
	binanceService.binanceClient = binance.NewClient(binanceService.apiKey, binanceService.apiSecret)
	return &binanceService
}

func init() {
	cwd, _ := os.Getwd()
	var dir string
	dir = os.Getenv("CONF_FILE")
	if dir == "" {
		dir = "/conf.env"
	}
	_ = godotenv.Load(cwd + dir)
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

	// Get pairInfo to correct quantity price deflections
	pairInfo := binanceService.GetPairInfo(pair)
	stepString := strconv.FormatFloat(pairInfo.StepSize, 'f', -1, 64)
	stepDecLength := len(stepString) - 2
	stepDecLengthFormatString := fmt.Sprintf("%%.%df", stepDecLength)
	quantityString := fmt.Sprintf(stepDecLengthFormatString, quantity)
	quantity, _ = strconv.ParseFloat(quantityString, 64)

	//Convert techan orderSide to binance SideType
	var sideType binance.SideType
	if orderSide == models.BUY {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	//Convert models orderType to binance orderType
	binanceOrderType := binance.OrderType(orderType)

	formatString := fmt.Sprintf("%%.%df", pairInfo.Precision)
	preparedOrder := binanceService.binanceClient.NewCreateOrderService().Symbol(pair).
		Side(sideType).Type(binanceOrderType).Quantity(fmt.Sprintf(formatString, quantity))

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
	defer func() {
		if r := recover(); r != nil {
			helpers.Logger.Errorln(fmt.Sprintf("Recovered. Error on DepthMonitor: (pair %s): %v, ", pair, r))
			time.Sleep(1 * time.Second)
			binanceService.DepthMonitor(pair, marketSnapshotsRecord)
		}
	}()
	binanceService.marketSnapshotsRecord = marketSnapshotsRecord
	binanceService.pair = pair

	doneC, _, err := binance.WsDepthServe(pair, binanceService.wsDepthHandler, binanceService.wsDepthErrHandler)
	if err != nil {
		helpers.Logger.Errorln(err)
		return
	}
	<-doneC
}

func (binanceService *BinanceService) TimeSeriesMonitor(pair, interval string, timeSeries *techan.TimeSeries, active *bool) {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logger.Errorln(fmt.Sprintf("Recovered. Error on TimeSeriesMonitor (pair %s): %v, ", pair, r))
			time.Sleep(1 * time.Second)
			binanceService.TimeSeriesMonitor(pair, interval, timeSeries, active)
		}
	}()
	binanceService.timeSeries = timeSeries
	binanceService.pair = pair
	binanceService.interval = interval
	binanceService.active = active

	klines, err := binanceService.binanceClient.NewKlinesService().Symbol(binanceService.pair).
		Interval(interval).Do(context.Background())
	if err != nil {
		helpers.Logger.Errorln("error getting klines: " + err.Error())
	}

	for _, k := range klines {
		//TODO: INTERVAL NOT STATIC!!
		period := techan.NewTimePeriod(time.Unix(k.OpenTime/1000, 0), time.Minute*60)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(k.Open)
		candle.ClosePrice = big.NewFromString(k.Close)
		candle.MaxPrice = big.NewFromString(k.High)
		candle.MinPrice = big.NewFromString(k.Low)
		candle.TradeCount = uint(k.TradeNum)
		candle.Volume = big.NewFromString(k.Volume)
		binanceService.dBService.AddOrUpdateCandle(*candle, binanceService.pair)
		binanceService.timeSeries.AddCandle(candle)
	}

	doneC, done, err := binance.WsKlineServe(binanceService.pair, interval, binanceService.wsKlineHandler, binanceService.wsKlineErrHandler)
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

	intervalSeconds := helpers.StringIntervalToSeconds(interval)

	provisionalLimit := limit % 1000
	if provisionalLimit == 0 {
		provisionalLimit = 1000
	}
	var startTime int64
	var resultKlines []*binance.Kline
	for iterations := 0; limit != 0; iterations++ {
		startTime = time.Now().Unix() - int64(intervalSeconds)*int64(limit)
		klines, err := binanceService.binanceClient.NewKlinesService().Symbol(pair).
			Interval(interval).Limit(provisionalLimit).StartTime(startTime * 1000).Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return timeSeries, err
		}

		for _, k := range klines {
			resultKlines = append(resultKlines, k)
		}

		limit -= provisionalLimit
		provisionalLimit = 1000
	}

	for _, k := range resultKlines {
		period := techan.NewTimePeriod(time.Unix(k.OpenTime/1000, 0), time.Minute*15)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(k.Open)
		candle.ClosePrice = big.NewFromString(k.Close)
		candle.MaxPrice = big.NewFromString(k.High)
		candle.MinPrice = big.NewFromString(k.Low)
		candle.TradeCount = uint(k.TradeNum)
		candle.Volume = big.NewFromString(k.Volume)
		binanceService.dBService.AddOrUpdateCandle(*candle, binanceService.pair)
		timeSeries.AddCandle(candle)
	}

	return timeSeries, nil
}

func (binanceService *BinanceService) GetMarkets(coin string, whitelist []string, blacklist []string) []string {
	var pairList []string

	blacklistStringify := strings.Join(blacklist, ",")
	whitelistStringify := strings.Join(whitelist, ",")

	info, _ := binanceService.binanceClient.NewExchangeInfoService().Do(context.Background())
	for _, symbol := range info.Symbols {

		if strings.Contains(symbol.Symbol, coin) &&
			(len(blacklist) == 0 || (len(blacklist) > 0 && !strings.Contains(blacklistStringify, symbol.Symbol))) &&
			(len(whitelist) == 0 || (len(whitelist) > 0 && strings.Contains(whitelistStringify, symbol.Symbol))) {
			pairList = append(pairList, symbol.Symbol)
		}
	}
	return pairList
}

func (binanceService *BinanceService) GetPairInfo(pair string) *models.PairInfo {
	info, _ := binanceService.binanceClient.NewExchangeInfoService().Do(context.Background())
	for _, symbol := range info.Symbols {
		if strings.Contains(symbol.Symbol, pair) {

			maxPrice, _ := strconv.ParseFloat(symbol.LotSizeFilter().MaxQuantity, 64)
			minPrice, _ := strconv.ParseFloat(symbol.LotSizeFilter().MinQuantity, 64)
			tickSize, _ := strconv.ParseFloat(symbol.LotSizeFilter().StepSize, 64)
			pairInfo := models.NewPairInfo(maxPrice, minPrice,
				tickSize, symbol.QuotePrecision)

			return pairInfo
		}
	}
	return nil
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
	binanceService.dBService.AddOrUpdateCandle(*candle, binanceService.pair)

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
		binanceService.wsDepthErrHandler(err)
		return
	}

	// Grab the record only if there are 2 bid and ask groups
	if marketSnapshot.WsDepthEvent != nil && marketSnapshot.WsDepthEvent.Bids != nil && marketSnapshot.WsDepthEvent.Asks != nil &&
		len(marketSnapshot.WsDepthEvent.Bids) > 2 && len(marketSnapshot.WsDepthEvent.Asks) > 2 {

		//Append to MarkerSnapshotsRecord
		reverseAny(*binanceService.marketSnapshotsRecord)
		*binanceService.marketSnapshotsRecord = append(*binanceService.marketSnapshotsRecord, marketSnapshot)
		reverseAny(*binanceService.marketSnapshotsRecord)

		if len(*binanceService.marketSnapshotsRecord) > 1020 {
			remove(*binanceService.marketSnapshotsRecord, 0)
		}

	}
}

func (binanceService *BinanceService) wsDepthErrHandler(err error) {
	helpers.Logger.Errorln("Error in Binace Depth monitor on pair " + binanceService.pair + ": " + err.Error())
	binanceService.DepthMonitor(binanceService.pair, binanceService.marketSnapshotsRecord)
}

func (binanceService *BinanceService) wsKlineErrHandler(err error) {
	helpers.Logger.Errorln("Error in Binace Kline monitor on pair " + binanceService.pair + ": " + err.Error())
	binanceService.TimeSeriesMonitor(binanceService.pair, binanceService.interval, binanceService.timeSeries, binanceService.active)
}

func reverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func remove(slice []models.MarketDepth, s int) []models.MarketDepth {
	return append(slice[:s], slice[s+1:]...)
}
