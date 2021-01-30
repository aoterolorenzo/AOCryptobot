package service

import (
	"../../model"
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"log"
	"math"
	"os"
	"strconv"
)

type MarketService struct {
	binanceClient    *binance.Client
	apiKey           string
	apiSecret        string
	pair             string
	MarketStatusList []model.MarketStatus
	Analytics        MarketAnalytics
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/marketmaker/service/binance/conf.env")
	if err != nil {
		log.Fatal("Error loading go.env file", err)
	}
}

func (marketService *MarketService) SetPair(pair string) {
	marketService.pair = pair
}

func (marketService *MarketService) ConfigureClient() {
	marketService.Analytics = MarketAnalytics{}
	marketService.apiKey = os.Getenv("apiKey")
	marketService.apiSecret = os.Getenv("apiSecret")
	marketService.binanceClient = binance.NewClient(marketService.apiKey, marketService.apiSecret)
}

func (marketService *MarketService) StartMonitor() {
	go marketService.monitor()
}

// Adds a model.MarketStatus to the record's model.MarketStatusList
func (marketService *MarketService) AppendStatus(marketStatus *model.MarketStatus) {
	reverseAny(marketService.MarketStatusList)
	marketService.MarketStatusList = append(marketService.MarketStatusList, *marketStatus)
	reverseAny(marketService.MarketStatusList)
}

func (marketService *MarketService) GetTotalBalance(asset string) (float64, error) {
	res, err := marketService.binanceClient.NewGetAccountService().Do(context.Background())
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

func (marketService *MarketService) MakeOrder(quantity float64, rate float64,
	sideType binance.SideType) (int64, error) {

	if sideType == binance.SideTypeBuy {
		quantity = quantity / rate
	}

	order, err := marketService.binanceClient.NewCreateOrderService().Symbol(marketService.pair).
		Side(sideType).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(fmt.Sprintf("%.2f", math.Floor(quantity*100)/100)).
		Price(fmt.Sprintf("%.2f", math.Floor(rate*100)/100)).Do(context.Background())

	if err != nil {
		return 0, err
	}

	return order.OrderID, nil
}

func (marketService *MarketService) MakeOCOOrder(quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
	sideType binance.SideType) (*binance.CreateOCOResponse, error) {

	if sideType == binance.SideTypeBuy {
		quantity = quantity / rate
	}

	order, err := marketService.binanceClient.NewCreateOCOService().Symbol(marketService.pair).Side(sideType).
		Price(fmt.Sprintf("%.2f", math.Floor(rate*100)/100)).
		StopPrice(fmt.Sprintf("%.2f", math.Floor(stopPrice*100)/100)).
		StopLimitPrice(fmt.Sprintf("%.2f", math.Floor(stopLimitPrice*100)/100)).
		Quantity(fmt.Sprintf("%.2f", math.Floor(quantity*100)/100)).
		StopLimitTimeInForce("GTC").Do(context.Background())

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (marketService *MarketService) GetOrder(orderId int64) (*binance.Order, error) {
	order, err := marketService.binanceClient.NewGetOrderService().Symbol(marketService.pair).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (marketService *MarketService) CancelOrder(orderId int64) error {
	_, err := marketService.binanceClient.NewCancelOrderService().Symbol(marketService.pair).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (marketService *MarketService) GetOrderStatus(orderId int64) (*binance.Order, error) {
	order, err := marketService.binanceClient.NewGetOrderService().OrderID(orderId).
		Symbol(marketService.pair).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (marketService *MarketService) monitor() {

	doneC, _, err := binance.WsDepthServe(marketService.pair, marketService.wsDepthHandler,
		marketService.errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneC
}

func (marketService *MarketService) wsDepthHandler(event *binance.WsDepthEvent) {
	marketStatus := model.MarketStatus{}
	err := marketStatus.Set(event)
	if err != nil {
		marketService.errHandler(err)
		return
	}

	marketService.AppendStatus(&marketStatus)

}

func (marketService *MarketService) errHandler(err error) {
	fmt.Println(err)
}
