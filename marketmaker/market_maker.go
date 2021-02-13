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
	"sync"
	"time"
)

const (
	MONITOR = "MONITOR"
	BUYING  = "BUYING"
	HOLDING = "HOLDING"
)

type MMStrategy struct {
	BinanceService       *exchangeService.BinanceService
	MarketService        *common.MarketService
	WalletService        *common.WalletService
	OrderBookService     *common.OrderBookService
	logList              *[]string
	logListMutex         *sync.Mutex
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
	threadName           string

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

func (m *MMStrategy) SetLogListAndMutex(logList *[]string, mutex *sync.Mutex) {
	m.logListMutex = mutex
	m.logList = logList
}

func (m *MMStrategy) SetThreadName(threadName string) {
	m.threadName = threadName
}

func (m *MMStrategy) Execute(waitTime int) {

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
	m.logListMutex.Lock()
	*m.logList = append(*m.logList, m.threadName+": Loading window data...")
	m.logListMutex.Unlock()

	for {
		if len(m.MarketService.MarketSnapshotsRecord) > m.monitorWindow {
			break
		}
	}

	time.Sleep(time.Duration(waitTime) * time.Second)

	m.logListMutex.Lock()
	*m.logList = append(*m.logList, m.threadName+": Data loaded, thread started")
	m.logListMutex.Unlock()

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

	lastPricePercentile, err := m.MarketService.CurrentPricePercentile(m.monitorWindow, &m.MarketService.MarketSnapshotsRecord)
	if err != nil {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+" e1:"+err.Error())
		m.logListMutex.Unlock()
		return
	}
	pctVariation, err := m.MarketService.PctVariation(m.monitorWindow, &m.MarketService.MarketSnapshotsRecord)
	if err != nil {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+" e2:"+err.Error())
		m.logListMutex.Unlock()
		return
	}

	if lastPricePercentile < m.topPercentileToTrade && lastPricePercentile > m.topPercentileToTrade {
		//fmt.Printf("Mercado fuera de percentil recomendado para iniciar compra %.2f", lastPricePercentile)
		return
	}

	if pctVariation < m.panicModeMargin {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+": Panic mode: Market going down")
		*m.logList = append(*m.logList, m.threadName+": Waiting for market...")
		m.logListMutex.Unlock()

		for {
			lastPricePercentile, err = m.MarketService.CurrentPricePercentile(m.monitorWindow, &m.MarketService.MarketSnapshotsRecord)
			if err != nil {
				m.logListMutex.Lock()
				*m.logList = append(*m.logList, m.threadName+" e1: "+err.Error())
				m.logListMutex.Unlock()
				return
			}
			pctVariation, err = m.MarketService.PctVariation(m.monitorWindow, &m.MarketService.MarketSnapshotsRecord)
			if err != nil {
				m.logListMutex.Lock()
				*m.logList = append(*m.logList, m.threadName+" e2: "+err.Error())
				m.logListMutex.Unlock()
				return
			}

			time.Sleep(1 * time.Second)

			if pctVariation < m.panicModeMargin {
				break
			}
		}
		return
	}

	m.logListMutex.Lock()
	*m.logList = append(*m.logList, m.threadName+": Time to buy")
	m.logListMutex.Unlock()

	m.buyRate = m.MarketService.CurrentPrice(&m.MarketService.MarketSnapshotsRecord) * (1 - m.buyMargin)
	balanceA, _ := m.BinanceService.GetTotalBalance(m.WalletService.Coin2)
	m.buyAmount = balanceA * m.pctAmountToTrade / 100

	buyOrder, err := m.BinanceService.MakeOrder(m.buyAmount, m.buyRate, binance.SideTypeBuy)
	if err != nil {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+" e3: "+err.Error())
		m.logListMutex.Unlock()
		return
	}
	m.buyOrder = buyOrder

	m.logListMutex.Lock()
	*m.logList = append(*m.logList,
		fmt.Sprintf(m.threadName+": Buy order emitted: rate %4f %s, amount %4f %s", m.buyRate, m.WalletService.Coin2,
			m.buyAmount, m.WalletService.Coin2))
	m.logListMutex.Unlock()

	m.state.Time = int(time.Now().Unix())

	m.OrderBookService.AddOpenOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))
	m.state.Current = BUYING

}

func (m *MMStrategy) buying() {

	orderStatus, err := m.BinanceService.GetOrderStatus(m.buyOrder.OrderID)
	if err != nil {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+" e4: "+err.Error())
		m.logListMutex.Unlock()
		return
	}

	if orderStatus.Status == binance.OrderStatusTypeFilled {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+":  Buy order filled")
		m.logListMutex.Unlock()
		m.OrderBookService.RemoveOpenOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))
		m.OrderBookService.AddFilledOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))

		//INICIAMOS ORDEN DE VENTA
		m.sellRate = m.buyRate * (1 + m.buyMargin) * (1 + m.sellMargin)
		m.sellAmount = m.buyAmount / m.buyRate
		m.stopPrice = m.sellRate * (1 - m.stopLossPct)
		m.stopLimitPrice = m.stopPrice * (1 + m.trailingStopLossPct) * (1 - 0.0004)

		m.sellOCOOrder, err = m.BinanceService.MakeOCOOrder(m.sellAmount, m.sellRate, m.stopPrice, m.stopLimitPrice, binance.SideTypeSell)
		if err != nil {
			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+" e5: "+err.Error())
			m.logListMutex.Unlock()
			return
		}

		m.logListMutex.Lock()
		*m.logList = append(*m.logList, fmt.Sprintf(m.threadName+":  Sell OCO order emmitted: rate %f %s, "+
			"cantidad %f %s, stop-loss %f %s ", m.sellRate, m.WalletService.Coin2, m.sellAmount, m.WalletService.Coin2,
			m.stopPrice, m.WalletService.Coin2))
		m.logListMutex.Unlock()
		m.OrderBookService.AddOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
		m.state.Current = HOLDING
		m.state.Time = int(time.Now().Unix())

	} else if m.state.Time+m.buyingTimeout < int(time.Now().Unix()) {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+": Buy timeout. Order canceled")
		m.logListMutex.Unlock()
		err = m.BinanceService.CancelOrder(m.buyOrder.OrderID)
		if err != nil {
			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+" e6: "+err.Error())
			m.logListMutex.Unlock()
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
			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+": STOP LOSS activated. Sold by strategy")
			m.logListMutex.Unlock()
			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.OrderBookService.AddFilledOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.state.Current = MONITOR
			m.state.Time = int(time.Now().Unix())
		} else if tempSellOrder.Status == binance.OrderStatusTypeFilled {
			// LIMIT FILLED
			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+": Sell successfully filled")
			m.logListMutex.Unlock()
			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.OrderBookService.AddFilledOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.state.Current = MONITOR
			m.state.Time = int(time.Now().Unix())
		}

	}

	// CHECK SELL IS NOT TIMEOUT init + 2 dias = ahora
	if m.sellingTimeout != 0 && m.state.Time+m.sellingTimeout < int(time.Now().Unix()) {
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+": Sell timeout")
		m.logListMutex.Unlock()

		orderA, err := m.BinanceService.GetOrder(m.sellOCOOrder.Orders[0].OrderID)
		if err != nil {
			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+" e7: "+err.Error())
			m.logListMutex.Unlock()
			return
		}

		orderB, err := m.BinanceService.GetOrder(m.sellOCOOrder.Orders[1].OrderID)
		if err != nil {
			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+" e8: "+err.Error())
			m.logListMutex.Unlock()
			return
		}

		if !(orderA.Status == binance.OrderStatusTypePartiallyFilled) &&
			!(orderB.Status == binance.OrderStatusTypePartiallyFilled) {
			err = m.BinanceService.CancelOrder(orderA.OrderID)
			if err != nil {
				m.logListMutex.Lock()
				*m.logList = append(*m.logList, m.threadName+" e9: "+err.Error())
				m.logListMutex.Unlock()
				return
			}

			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+": Sell OCO order canceled")
			*m.logList = append(*m.logList, m.threadName+": Sell order at market price emitted")
			m.logListMutex.Unlock()

			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))

			order, err := m.BinanceService.MakeOrder(m.sellAmount,
				m.MarketService.MarketSnapshotsRecord[0].HigherBidPrice, binance.SideTypeSell)
			m.OrderBookService.AddOpenOrder(m.BinanceService.OrderResponseToOrder(*order))
			if err != nil {
				m.logListMutex.Lock()
				*m.logList = append(*m.logList, m.threadName+" e10: "+err.Error())
				m.logListMutex.Unlock()
				return
			}

			m.logListMutex.Lock()
			*m.logList = append(*m.logList, m.threadName+": Waiting to fill sell timeout order")
			m.logListMutex.Unlock()
			for {
				timeoutSellORder, err := m.BinanceService.GetOrder(order.OrderID)
				if err != nil {
					m.logListMutex.Lock()
					*m.logList = append(*m.logList, m.threadName+" e11: "+err.Error())
					m.logListMutex.Unlock()
					return
				}

				if timeoutSellORder.Status == binance.OrderStatusTypeFilled {
					m.logListMutex.Lock()
					*m.logList = append(*m.logList, m.threadName+": Sell timeout order filled")
					m.logListMutex.Unlock()
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
