package binance

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"os"
	"strconv"
)

var logger = helpers.Logger{}

type BinanceService struct {
	binanceClient *binance.Client
	apiKey        string
	apiSecret     string
	pair          string
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/services/binance/conf.env")
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

func (binanceService *BinanceService) MakeOrder(quantity float64, rate float64, sideType binance.SideType) (*binance.CreateOrderResponse, error) {

	if sideType == binance.SideTypeBuy {
		quantity = quantity / rate
	}

	order, err := binanceService.binanceClient.NewCreateOrderService().Symbol(binanceService.pair).
		Side(sideType).Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(fmt.Sprintf("%.5f", quantity)).
		Price(fmt.Sprintf("%.2f", rate)).Do(context.Background())

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (binanceService *BinanceService) MakeOCOOrder(quantity float64, rate float64, stopPrice float64, stopLimitPrice float64,
	sideType binance.SideType) (*binance.CreateOCOResponse, error) {

	if sideType == binance.SideTypeBuy {
		quantity = quantity / rate
	}

	order, err := binanceService.binanceClient.NewCreateOCOService().Symbol(binanceService.pair).Side(sideType).
		Price(fmt.Sprintf("%.2f", rate)).
		StopPrice(fmt.Sprintf("%.2f", stopPrice)).
		StopLimitPrice(fmt.Sprintf("%.2f", stopLimitPrice)).
		Quantity(fmt.Sprintf("%.5f", quantity)).
		StopLimitTimeInForce("GTC").Do(context.Background())

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (binanceService *BinanceService) GetOrder(orderId int64) (*binance.Order, error) {
	order, err := binanceService.binanceClient.NewGetOrderService().Symbol(binanceService.pair).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (binanceService *BinanceService) CancelOrder(orderId int64) error {
	_, err := binanceService.binanceClient.NewCancelOrderService().Symbol(binanceService.pair).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (binanceService *BinanceService) GetOrderStatus(orderId int64) (*binance.Order, error) {
	order, err := binanceService.binanceClient.NewGetOrderService().OrderID(orderId).
		Symbol(binanceService.pair).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return order, nil
}

//func (binanceService *BinanceService) monitor() {
//
//	doneC, _, err := binance.WsDepthServe(binanceService.pair, binanceService.wsDepthHandler,
//		binanceService.errHandler)
//	if err != nil {
//		log.Fatal(err)
//		return
//	}
//	<-doneC
//}

func (binanceService *BinanceService) WsDepth(dh binance.WsDepthHandler, eh binance.ErrHandler) {
	//TODO: Use exchange logic from here and not on market services
	doneC, _, err := binance.WsDepthServe(binanceService.pair, dh, eh)
	if err != nil {
		logger.Errorln(err)
		return
	}
	<-doneC
}

func (binanceService *BinanceService) OrderResponseToOrder(o binance.CreateOrderResponse) binance.Order {
	return binance.Order{
		Symbol:                   o.Symbol,
		OrderID:                  o.OrderID,
		ClientOrderID:            o.ClientOrderID,
		Price:                    o.Price,
		OrigQuantity:             o.OrigQuantity,
		ExecutedQuantity:         o.ExecutedQuantity,
		CummulativeQuoteQuantity: o.CummulativeQuoteQuantity,
		Status:                   o.Status,
		TimeInForce:              o.TimeInForce,
		Type:                     o.Type,
		Side:                     o.Side,
		StopPrice:                "",
		IcebergQuantity:          "",
		Time:                     o.TransactTime,
		UpdateTime:               0,
		IsWorking:                false,
	}
}

func (binanceService *BinanceService) OCOOrderResponseToOrder(o binance.CreateOCOResponse) binance.Order {
	for _, m := range o.Orders {
		return binance.Order{
			Symbol:                   m.Symbol,
			OrderID:                  m.OrderID,
			ClientOrderID:            m.ClientOrderID,
			Price:                    "",
			OrigQuantity:             "",
			ExecutedQuantity:         "",
			CummulativeQuoteQuantity: "",
			Status:                   "",
			TimeInForce:              "",
			Type:                     "OCO",
			Side:                     "SELL",
			StopPrice:                "",
			IcebergQuantity:          "",
			Time:                     0,
			UpdateTime:               0,
			IsWorking:                false,
		}
	}
	return binance.Order{}
}
