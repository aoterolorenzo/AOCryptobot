package marketmaker

import (
	"../marketmaker/helpers"
	exchangeService "../marketmaker/service/binance"
	"../marketmaker/service/common"
	"fmt"
	"github.com/adshao/go-binance/v2"
	tm "github.com/buger/goterm"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	MONITOR = "MONITOR"
	BUYING  = "BUYING"
	HOLDING = "HOLDING"
)

const (
	FILLED   = "FILLED"
	CANCELED = "CANCELED"
	NEXT     = "NEXT"
)

type MMStrategy struct {
	BinanceService       *exchangeService.BinanceService
	MarketService        *common.MarketService
	WalletService        *common.WalletService
	OrderBookService     *common.OrderBookService
	logList              *[]string
	state                helpers.STATE
	startCoin1Amout      float64
	startCoin2Amout      float64
	monitorWindow        int
	pctAmountToTrade     float64
	buyMargin            float64
	sellMargin           float64
	stopLossPct          float64
	trailingStopLossPct  float64
	topPercentileToTrade float64
	lowPercentileToTrade float64
	panicModeMargin      float64
	monitorFrequency     int
	buyingTimeout        int
	sellingTimeout       int

	// Execution time variables
	buyRate   float64
	buyAmount float64
	buyOrder  *binance.CreateOrderResponse

	sellRate       float64
	sellAmount     float64
	sellOCOOrder   *binance.CreateOCOResponse
	sellOrder      *binance.CreateOrderResponse
	stopPrice      float64
	stopLimitPrice float64
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/marketmaker/strategy.env")
	if err != nil {
		log.Fatal("Error loading go.env file", err)
	}
}

func (m *MMStrategy) SetServices(binanceService *exchangeService.BinanceService, marketService *common.MarketService,
	walletService *common.WalletService, orderBookService *common.OrderBookService) {
	m.BinanceService = binanceService
	m.MarketService = marketService
	m.WalletService = walletService
	m.OrderBookService = orderBookService
}

func (m *MMStrategy) SetLogList(logList *[]string) {
	m.logList = logList
}

func (m *MMStrategy) Execute() {

	// Get .env file strategy variables
	m.monitorWindow, _ = strconv.Atoi(os.Getenv("monitorWindow"))
	m.pctAmountToTrade, _ = strconv.ParseFloat(os.Getenv("pctAmountToTrade"), 64)
	m.buyMargin, _ = strconv.ParseFloat(os.Getenv("buyMargin"), 64)
	m.sellMargin, _ = strconv.ParseFloat(os.Getenv("sellMargin"), 64)
	m.stopLossPct, _ = strconv.ParseFloat(os.Getenv("stopLossPct"), 64)
	m.topPercentileToTrade, _ = strconv.ParseFloat(os.Getenv("topPercentileToTrade"), 64)
	m.lowPercentileToTrade, _ = strconv.ParseFloat(os.Getenv("lowPercentileToTrade"), 64)
	m.panicModeMargin, _ = strconv.ParseFloat(os.Getenv("panicModeMargin"), 64)
	m.monitorFrequency, _ = strconv.Atoi(os.Getenv("monitorFrequency"))
	m.buyingTimeout, _ = strconv.Atoi(os.Getenv("buyingTimeout"))
	m.sellingTimeout, _ = strconv.Atoi(os.Getenv("sellingTimeout"))
	m.trailingStopLossPct, _ = strconv.ParseFloat(os.Getenv("trailingStopLossPct"), 64)

	// Set initial state to MONITOR and start marketService monitor
	m.state.Current = MONITOR

	// Wait window iterations until monitor loads
	*m.logList = append(*m.logList, "Loading window data...")

	for {
		if len(m.MarketService.MarketSnapshotsRecord) > m.monitorWindow {
			break
		}
	}

	*m.logList = append(*m.logList, "Data loaded, thread started")

	for {
		// Switch functions depending on the current state
		tm.Clear()
		switch m.state.Current {
		case MONITOR:
			m.monitor()
		case BUYING:
			m.buying()
		case HOLDING:
			m.holding()
		}

		// Wait frequency to repeat
		time.Sleep(time.Duration(m.monitorFrequency) * time.Second)
	}

}

func (m *MMStrategy) monitor() {

	lastPricePercentile, _ := m.MarketService.CurrentPricePercentile(m.monitorWindow, &m.MarketService.MarketSnapshotsRecord)
	pctVariation, _ := m.MarketService.PctVariation(m.monitorWindow, &m.MarketService.MarketSnapshotsRecord)

	if lastPricePercentile < m.topPercentileToTrade && lastPricePercentile > m.topPercentileToTrade {
		//fmt.Printf("Mercado fuera de percentil recomendado para iniciar compra %.2f", lastPricePercentile)
		return
	}

	if pctVariation < m.panicModeMargin {
		*m.logList = append(*m.logList, "Panic mode: Market going down.")
		return
	}

	*m.logList = append(*m.logList, "Strategy: Time to buy.")

	m.buyRate = m.MarketService.CurrentPrice(&m.MarketService.MarketSnapshotsRecord) * (1 - m.buyMargin)
	balanceA, _ := m.BinanceService.GetTotalBalance("EUR")
	m.buyAmount = balanceA * m.pctAmountToTrade / 100

	//if m.startCoin1Amout > balanceA {
	//	*m.logList = append(*m.logList, "SALIENDO!!! DETECTADAS PERDIDAS!!")
	//	os.Exit(1)
	//}

	buyOrder, err := m.BinanceService.MakeOrder(m.buyAmount, m.buyRate, binance.SideTypeBuy)
	if err != nil {
		println(1)
		*m.logList = append(*m.logList, "e3:"+err.Error())
		return
	}
	m.buyOrder = buyOrder

	*m.logList = append(*m.logList,
		fmt.Sprintf("Strategy: Buy order emitted: rate %4f EUR, amount %4f EUR", m.buyRate, m.buyAmount))

	m.state.Time = int(time.Now().Unix())

	m.OrderBookService.AddOpenOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))
	m.state.Current = BUYING

}

func (m *MMStrategy) buying() {

	orderStatus, err := m.BinanceService.GetOrderStatus(m.buyOrder.OrderID)
	if err != nil {
		return
	}

	if orderStatus.Status == binance.OrderStatusTypeFilled {
		*m.logList = append(*m.logList, "Strategy:  Buy order filled")
		m.OrderBookService.RemoveOpenOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))
		m.OrderBookService.AddFilledOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))

		//INICIAMOS ORDEN DE VENTA
		m.sellRate = m.buyRate * (1 + m.buyMargin) * (1 + m.sellMargin)
		m.sellAmount = m.buyAmount / m.buyRate
		m.stopPrice = m.sellRate * (1 - m.stopLossPct)
		m.stopLimitPrice = m.stopPrice * (1 + m.trailingStopLossPct) * (1 - 0.0004)

		m.sellOCOOrder, err = m.BinanceService.MakeOCOOrder(m.sellAmount, m.sellRate, m.stopPrice, m.stopLimitPrice, binance.SideTypeSell)
		if err != nil {
			println(5)
			*m.logList = append(*m.logList, err.Error())
		}

		*m.logList = append(*m.logList, fmt.Sprintf("Strategy:  Sell OCO order emmitted: rate %f EUR, "+
			"cantidad %f EUR, stop-loss %f EUR ", m.sellRate, m.sellAmount, m.stopPrice))
		m.OrderBookService.AddOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
		m.state.Current = HOLDING
		m.state.Time = int(time.Now().Unix())

	} else if m.state.Time+m.buyingTimeout < int(time.Now().Unix()) {
		*m.logList = append(*m.logList, "Strategy: Buy timeout. Order canceled")
		err = m.BinanceService.CancelOrder(m.buyOrder.OrderID)
		if err != nil {
			*m.logList = append(*m.logList, "cancel order"+err.Error())
			return
		}

		m.OrderBookService.RemoveOpenOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))

		m.state.Current = MONITOR
		m.state.Time = int(time.Now().Unix())
	}

}

func (m *MMStrategy) holding() {

	for _, order := range m.sellOCOOrder.Orders {
		// GET FIRST ORDER
		tempSellOrder, err := m.BinanceService.GetOrder(order.OrderID)
		if err != nil {
			return
		}

		if tempSellOrder.Status == binance.OrderStatusTypeFilled &&
			tempSellOrder.Type == binance.OrderTypeStopLossLimit {
			// STOP LOSS LIMIT FILLED
			*m.logList = append(*m.logList, "Strategy: STOP LOSS activated. Sold by strategy")
			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.OrderBookService.AddFilledOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.state.Current = MONITOR
			m.state.Time = int(time.Now().Unix())
		} else if tempSellOrder.Status == binance.OrderStatusTypeFilled {
			// LIMIT FILLED
			*m.logList = append(*m.logList, "Strategy: Sell successfully filled")
			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.OrderBookService.AddFilledOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.state.Current = MONITOR
			m.state.Time = int(time.Now().Unix())
		}

	}

	// CHECK SELL IS NOT TIMEOUT init + 2 dias = ahora
	if m.sellingTimeout != 0 && m.state.Time+m.sellingTimeout < int(time.Now().Unix()) {
		*m.logList = append(*m.logList, "Strategy: Sell timeout")

		orderA, err := m.BinanceService.GetOrder(m.sellOCOOrder.Orders[0].OrderID)
		if err != nil {
			println(7)
			*m.logList = append(*m.logList, "e1: "+err.Error())
		}

		orderB, err := m.BinanceService.GetOrder(m.sellOCOOrder.Orders[1].OrderID)
		if err != nil {
			*m.logList = append(*m.logList, "get order: "+err.Error())
			m.holding()
		}

		if !(orderA.Status == binance.OrderStatusTypePartiallyFilled) &&
			!(orderB.Status == binance.OrderStatusTypePartiallyFilled) {
			err = m.BinanceService.CancelOrder(orderA.OrderID)
			if err != nil {
				*m.logList = append(*m.logList, "cancel order:"+err.Error())
			}

			*m.logList = append(*m.logList, "Strategy: Sell OCO order canceled")
			*m.logList = append(*m.logList, "Strategy: Sell order at market price emitted")

			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))

			order, err := m.BinanceService.MakeOrder(m.sellAmount,
				m.MarketService.MarketSnapshotsRecord[0].HigherBidPrice, binance.SideTypeSell)
			m.OrderBookService.AddOpenOrder(m.BinanceService.OrderResponseToOrder(*order))
			if err != nil {
				println(11)
				*m.logList = append(*m.logList, err.Error())
			}

			*m.logList = append(*m.logList, "Strategy: Waiting to fill sell timeout order")
			for {
				timeoutSellORder, err := m.BinanceService.GetOrder(order.OrderID)
				if err != nil {
					println(12)
					*m.logList = append(*m.logList, "e2"+err.Error())
					break
				}

				if timeoutSellORder.Status == binance.OrderStatusTypeFilled {
					*m.logList = append(*m.logList, "Strategy: Sell timeout order filled")
					m.OrderBookService.RemoveOpenOrder(m.BinanceService.OrderResponseToOrder(*order))
					m.OrderBookService.AddFilledOrder(m.BinanceService.OrderResponseToOrder(*order))
					m.state.Current = MONITOR
					m.state.Time = int(time.Now().Unix())
					break
				}

				time.Sleep(1 * time.Second)
			}
		}
	}

}
