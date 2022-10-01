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
	"strconv"
	"strings"
	"time"
)

type PaperService struct {
	binanceClient *binance.Client
	timeSeries    *techan.TimeSeries
}

func NewPaperService() *PaperService {
	apiKey := os.Getenv("binanceAPIKey")
	apiSecret := os.Getenv("binanceAPISecret")
	binanceClient := binance.NewClient(apiKey, apiSecret)
	return &PaperService{
		binanceClient: binanceClient,
	}
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/conf.env")
	if err != nil {
		helpers.Logger.Errorln("Error loading go.env file", err)
	}
}

func (paperService *PaperService) GetTotalBalance(asset string) (float64, error) {
	return 3000, nil
}

func (paperService *PaperService) GetAvailableBalance(asset string) (float64, error) {
	return 10670.73, nil
}

func (paperService *PaperService) GetLockedBalance(asset string) (float64, error) {
	return 0, nil
}

func (paperService *PaperService) MakeOrder(pair string, quantity float64, rate float64,
	orderType models.OrderType, orderSide models.OrderSide) (models.Order, error) {

	var quantityString string
	var rateString string
	quantityString = fmt.Sprintf("%f", quantity)
	rateString = fmt.Sprintf("%f", rate)

	if orderSide == models.BUY {
		quantity /= rate
		quantityString = fmt.Sprintf("%f", quantity)
	}

	cumulativeQuantity := quantity * rate
	cumulativeQuantityString := fmt.Sprintf("%f", cumulativeQuantity)

	//Convert techan orderSide to binance SideType
	var sideType models.SideType
	if orderSide == models.BUY {
		sideType = models.SideTypeBuy
	} else {
		sideType = models.SideTypeSell
	}

	order := models.NewOrder(pair, 0, "0", rateString, quantityString, quantityString, cumulativeQuantityString, models.OrderStatusTypeFilled,
		orderType, sideType, time.Now().Unix(), time.Now().Unix(), false, false)

	return order, nil
}

func (paperService *PaperService) MakeOCOOrder(pair string, quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
	orderSide models.OrderSide) (models.OCOOrder, error) {
	return models.OCOOrder{}, nil
}

func (paperService *PaperService) GetOrder(order models.Order) (models.Order, error) {
	return models.NewEmptyOrder(), nil
}

func (paperService *PaperService) CancelOrder(order models.Order) error {
	return nil
}

func (paperService *PaperService) GetOrderStatus(order models.Order) (models.OrderStatusType, error) {
	return models.OrderStatusTypeFilled, nil
}

func (paperService *PaperService) DepthMonitor(pair string, marketSnapshotsRecord *[]models.MarketDepth) {
}

func (paperService *PaperService) TimeSeriesMonitor(pair string, interval string, timeSeries *techan.TimeSeries, active *bool) {
	paperService.timeSeries = timeSeries

	klines, err := paperService.binanceClient.NewKlinesService().Symbol(pair).
		Interval(interval).Do(context.Background())
	if err != nil {
		helpers.Logger.Errorln("error getting klines: " + err.Error())
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
		paperService.timeSeries.AddCandle(candle)
	}

	doneC, done, err := binance.WsKlineServe(pair, interval, paperService.wsKlineHandler, paperService.errHandler)
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

func (paperService *PaperService) GetSeries(pair string, interval string, limit int) (techan.TimeSeries, error) {
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
		klines, err := paperService.binanceClient.NewKlinesService().Symbol(pair).
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
		timeSeries.AddCandle(candle)
	}

	return timeSeries, nil
}

func (paperService *PaperService) GetMarkets(coin string, whitelist []string, blacklist []string) []string {
	var pairList []string

	blacklistStringify := strings.Join(blacklist, ",")
	whitelistStringify := strings.Join(whitelist, ",")

	info, _ := paperService.binanceClient.NewExchangeInfoService().Do(context.Background())
	for _, symbol := range info.Symbols {

		if strings.Contains(symbol.Symbol, coin) &&
			(len(blacklist) == 0 || (len(blacklist) > 0 && !strings.Contains(blacklistStringify, symbol.Symbol))) &&
			(len(whitelist) == 0 || (len(whitelist) > 0 && strings.Contains(whitelistStringify, symbol.Symbol))) {
			pairList = append(pairList, symbol.Symbol)
		}
	}
	return pairList
}

func (paperService *PaperService) GetPairInfo(pair string) *models.PairInfo {
	info, _ := paperService.binanceClient.NewExchangeInfoService().Do(context.Background())
	for _, symbol := range info.Symbols {
		if strings.Contains(symbol.Symbol, pair) {

			maxPrice, _ := strconv.ParseFloat(symbol.PriceFilter().MaxPrice, 64)
			minPrice, _ := strconv.ParseFloat(symbol.PriceFilter().MinPrice, 64)
			tickSize, _ := strconv.ParseFloat(symbol.PriceFilter().TickSize, 64)
			pairInfo := models.NewPairInfo(maxPrice, minPrice,
				tickSize, symbol.QuotePrecision)

			return pairInfo
		}
	}
	return nil
}

func (paperService *PaperService) wsKlineHandler(event *binance.WsKlineEvent) {
	lastCandle := paperService.timeSeries.Candles[len(paperService.timeSeries.Candles)-1]

	period := techan.NewTimePeriod(time.Unix(event.Kline.StartTime/1000, 0), time.Minute*15)
	candle := techan.NewCandle(period)
	candle.OpenPrice = big.NewFromString(event.Kline.Open)
	candle.ClosePrice = big.NewFromString(event.Kline.Close)
	candle.MaxPrice = big.NewFromString(event.Kline.High)
	candle.MinPrice = big.NewFromString(event.Kline.Low)
	candle.TradeCount = uint(event.Kline.TradeNum)
	candle.Volume = big.NewFromString(event.Kline.Volume)

	if lastCandle.Period != techan.NewTimePeriod(time.Unix(event.Kline.StartTime/1000, 0), time.Minute*15) {
		paperService.timeSeries.AddCandle(candle)
	} else if lastCandle.Period == techan.NewTimePeriod(time.Unix(event.Kline.StartTime/1000, 0), time.Minute*15) {
		paperService.timeSeries.Candles[len(paperService.timeSeries.Candles)-1] = candle
	}
}

func (paperService *PaperService) errHandler(err error) {
	helpers.Logger.Errorln(err)
}
