package marketmaker

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	tm "github.com/buger/goterm"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gitlab.com/aoterocom/AOCryptobot/marketmaker/helpers"
	"gitlab.com/aoterocom/AOCryptobot/marketmaker/services"
	exchangeService "gitlab.com/aoterocom/AOCryptobot/marketmaker/services/binance"
	"os"
	"strconv"
	"sync"
	"time"
)

var logger = helpers.Logger{}

const (
	MONITOR = "MONITOR"
	BUYING  = "BUYING"
	HOLDING = "HOLDING"
)

type MMStrategy struct {
	BinanceService       *exchangeService.BinanceService
	MarketService        *services.MarketService
	WalletService        *services.WalletService
	OrderBookService     *services.OrderBookService
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

	sellRate              float64
	sellAmount            float64
	sellOCOOrder          *binance.CreateOCOResponse
	sellOrder             *binance.CreateOrderResponse
	stopPrice             float64
	stopLimitPrice        float64
	afterSellingCoolDown  float64
	afterStopLossCoolDown float64
	threadNumber          int
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/marketmaker/strategy.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
}

func (m *MMStrategy) SetServices(binanceService *exchangeService.BinanceService, marketService *services.MarketService,
	walletService *services.WalletService, orderBookService *services.OrderBookService) {
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
	m.threadNumber, _ = strconv.Atoi(os.Getenv("threadNumber"))
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
	m.afterSellingCoolDown, _ = strconv.ParseFloat(os.Getenv("afterSellingCoolDown"), 64)
	m.afterStopLossCoolDown, _ = strconv.ParseFloat(os.Getenv("afterStopLossCoolDown"), 64)

	// Set initial state to MONITOR and start marketService monitor
	m.state.Current = MONITOR
	m.buyAmount = -1

	//Set some dispersion between threads
	m.monitorWindow += waitTime

	// Wait window iterations until monitor loads
	m.logAndList("Loading window data...", log.InfoLevel)

	for {
		if len(m.MarketService.MarketSnapshotsRecord) > m.monitorWindow {
			break
		}
	}

	time.Sleep(time.Duration(waitTime) * time.Second)

	m.logAndList("Data loaded, thread started", log.InfoLevel)

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

	lastPricePercentile, err := m.MarketService.CurrentPricePercentile(m.monitorWindow)
	if err != nil {
		m.logAndList(" e1:"+err.Error(), log.ErrorLevel)
		return
	}
	pctVariation, err := m.MarketService.PctVariation(m.monitorWindow)
	if err != nil {
		m.logAndList(" e2:"+err.Error(), log.ErrorLevel)
		return
	}

	if lastPricePercentile < m.topPercentileToTrade && lastPricePercentile > m.topPercentileToTrade {
		//fmt.Printf("Mercado fuera de percentil recomendado para iniciar compra %.2f", lastPricePercentile)
		return
	}

	if pctVariation < m.panicModeMargin {
		m.logAndList("Panic mode: Market going down", log.InfoLevel)
		m.logAndList("Waiting for market...", log.InfoLevel)

		for {
			lastPricePercentile, err = m.MarketService.CurrentPricePercentile(m.monitorWindow)
			if err != nil {
				m.logAndList("e1: "+err.Error(), log.ErrorLevel)
				return
			}
			pctVariation, err = m.MarketService.PctVariation(m.monitorWindow)
			if err != nil {
				m.logAndList("e2: "+err.Error(), log.ErrorLevel)
				return
			}

			time.Sleep(1 * time.Second)

			if pctVariation < m.panicModeMargin {
				break
			}
		}
		return
	}

	m.logAndList("Time to buy", log.InfoLevel)

	currentPrice := m.MarketService.CurrentPrice()
	m.buyRate = currentPrice * (1 - m.buyMargin)
	balanceA := m.WalletService.GetTotalAssetsBalance(currentPrice)
	m.buyAmount = balanceA * m.pctAmountToTrade / 100

	buyOrder, err := m.BinanceService.MakeOrder(m.buyAmount, m.buyRate, binance.SideTypeBuy)
	if err != nil {
		m.logAndList("e3: "+err.Error(), log.ErrorLevel)
		return
	}
	m.buyOrder = buyOrder

	m.logAndList(fmt.Sprintf("Buy order #%d emitted: rate %4f %s, amount %4f %s", m.buyOrder.OrderID, m.buyRate, m.WalletService.Coin2,
		m.buyAmount, m.WalletService.Coin2), log.InfoLevel)

	m.state.Time = int(time.Now().Unix())

	m.OrderBookService.AddOpenOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))
	m.state.Current = BUYING
	m.state.Time = int(time.Now().Unix())

}

func (m *MMStrategy) buying() {

	order, err := m.BinanceService.GetOrderStatus(m.buyOrder.OrderID)
	if err != nil {
		//m.logAndList("e4: "+err.Error(), log.ErrorLevel)
		return
	}

	if order.Status == binance.OrderStatusTypeFilled {
		m.logAndList(fmt.Sprintf("Buy order #%d filled", m.buyOrder.OrderID), log.InfoLevel)
		m.OrderBookService.RemoveOpenOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))
		m.OrderBookService.AddFilledOrder(m.BinanceService.OrderResponseToOrder(*m.buyOrder))

		err = m.WalletService.UpdateWallet()
		if err != nil {
			logger.Errorln("Error updating wallet" + err.Error())
			return
		}

		//INICIAMOS ORDEN DE VENTA
		m.sellRate = m.buyRate * (1 + m.buyMargin) * (1 + m.sellMargin)
		m.sellAmount, _ = strconv.ParseFloat(order.ExecutedQuantity, 64)
		m.stopPrice = m.sellRate * (1 - m.stopLossPct)
		m.stopLimitPrice = m.stopPrice * (1 - 0.00075)

		// Clean up the residues left by the decimals and broken thread amounts
		freeAsset, err := m.WalletService.GetFreeAssetBalance(m.WalletService.Coin1)
		if err != nil {
			logger.Errorln("Error getting asset: " + err.Error())
			return
		}

		// If there is residue or any broken thread had left a position on the other side, bring it on the OCO order
		if freeAsset < m.sellAmount+(m.sellAmount*0.6) || m.OrderBookService.OpenOrdersCount() == m.threadNumber-1 {
			m.sellAmount = freeAsset * 0.999
		}

		for {
			m.sellOCOOrder, err = m.BinanceService.MakeOCOOrder(m.sellAmount, m.sellRate, m.stopPrice, m.stopLimitPrice, binance.SideTypeSell)
			time.Sleep(500 * time.Millisecond)
			if err != nil {
				m.logAndList("e5: "+err.Error(), log.ErrorLevel)
				m.sellAmount *= 0.998
				m.logAndList(fmt.Sprintf("try : %.4f", m.sellAmount), log.ErrorLevel)
				continue
			}

			m.logAndList(fmt.Sprintf("Sell OCO order #%d/#%d emitted:", m.sellOCOOrder.Orders[0].OrderID,
				m.sellOCOOrder.Orders[1].OrderID), log.InfoLevel)
			m.logAndList(fmt.Sprintf("Rate %f %s, Quant %f %s, Stop-Loss %f %s ", m.sellRate,
				m.WalletService.Coin2, m.sellAmount, m.WalletService.Coin1, m.stopPrice, m.WalletService.Coin2), log.InfoLevel)
			m.OrderBookService.AddOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.state.Current = HOLDING
			m.state.Time = int(time.Now().Unix())
			break
		}

	} else if order.Status != binance.OrderStatusTypePartiallyFilled && m.state.Time+m.buyingTimeout < int(time.Now().Unix()) {
		m.logAndList(fmt.Sprintf("Buy timeout. Order #%d canceled", m.buyOrder.OrderID), log.InfoLevel)
		err = m.BinanceService.CancelOrder(m.buyOrder.OrderID)
		if err != nil {
			m.logAndList("e6: "+err.Error(), log.WarnLevel)
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
			// Stop-loss triggered
			m.logAndList(fmt.Sprintf("Sell order #%d/#%d filled by Stop-Loss", m.sellOCOOrder.Orders[0].OrderID,
				m.sellOCOOrder.Orders[1].OrderID), log.InfoLevel)
			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.OrderBookService.AddFilledOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			err = m.WalletService.UpdateWallet()

			// Stop-loss cooldown
			time.Sleep(time.Duration(m.afterStopLossCoolDown) * time.Second)

			if err != nil {
				logger.Errorln("Error updating wallet" + err.Error())
				return
			}
			m.state.Current = MONITOR
			m.state.Time = int(time.Now().Unix())
		} else if tempSellOrder.Status == binance.OrderStatusTypeFilled {
			// LIMIT FILLED
			m.logAndList(fmt.Sprintf("Sell order #%d/#%d successfully filled", m.sellOCOOrder.Orders[0].OrderID,
				m.sellOCOOrder.Orders[1].OrderID), log.InfoLevel)
			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			m.OrderBookService.AddFilledOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))
			err = m.WalletService.UpdateWallet()
			if err != nil {
				logger.Errorln("Error updating wallet" + err.Error())
				return
			}
			m.state.Current = MONITOR

			m.logAndList("Cooling down", log.InfoLevel)
			m.state.Time = int(time.Now().Unix())
			time.Sleep(time.Duration(m.afterSellingCoolDown) * time.Second)
		}

	}

	// CHECK SELL IS NOT TIMEOUT init + 2 dias = ahora
	if m.sellingTimeout != 0 && m.state.Time+m.sellingTimeout < int(time.Now().Unix()) {
		m.logAndList(fmt.Sprintf("Sell OCO order #%d/#%d timed out", m.sellOCOOrder.Orders[0].OrderID,
			m.sellOCOOrder.Orders[1].OrderID), log.InfoLevel)

		orderA, err := m.BinanceService.GetOrder(m.sellOCOOrder.Orders[0].OrderID)
		if err != nil {
			m.logAndList("e7: "+err.Error(), log.ErrorLevel)
			return
		}

		orderB, err := m.BinanceService.GetOrder(m.sellOCOOrder.Orders[1].OrderID)
		if err != nil {
			m.logAndList("e8: "+err.Error(), log.ErrorLevel)
			return
		}

		if !(orderA.Status == binance.OrderStatusTypePartiallyFilled) && !(orderA.Status == binance.OrderStatusTypeFilled) &&
			!(orderB.Status == binance.OrderStatusTypePartiallyFilled) && !(orderB.Status == binance.OrderStatusTypeFilled) {
			err = m.BinanceService.CancelOrder(orderA.OrderID)
			if err != nil {
				m.logAndList("e9: "+err.Error(), log.ErrorLevel)
				return
			}

			m.logAndList(fmt.Sprintf("Sell OCO order #%d/#%d canceled", m.sellOCOOrder.Orders[0].OrderID,
				m.sellOCOOrder.Orders[1].OrderID), log.InfoLevel)

			m.OrderBookService.RemoveOpenOrder(m.BinanceService.OCOOrderResponseToOrder(*m.sellOCOOrder))

			order, err := m.BinanceService.MakeOrder(m.sellAmount,
				m.MarketService.MarketSnapshotsRecord[0].HigherBidPrice*(1-0.0005), binance.SideTypeSell)
			if err != nil {
				m.logAndList("e10: "+err.Error(), log.ErrorLevel)
				return
			}
			m.OrderBookService.AddOpenOrder(m.BinanceService.OrderResponseToOrder(*order))
			m.logAndList(fmt.Sprintf("Sell order #%d emitted at market price", order.OrderID), log.InfoLevel)
			m.logAndList(fmt.Sprintf("Waiting to fill #%d sell timeout order", order.OrderID), log.InfoLevel)
			for {
				timeoutSellOrder, err := m.BinanceService.GetOrder(order.OrderID)
				if err != nil {
					m.logAndList(" e11: "+err.Error(), log.ErrorLevel)
					m.state.Current = MONITOR
					m.state.Time = int(time.Now().Unix())
					return
				}

				if timeoutSellOrder.Status == binance.OrderStatusTypeFilled {
					m.logAndList(fmt.Sprintf("Sell timeout order #%d  filled", timeoutSellOrder.OrderID), log.InfoLevel)
					m.OrderBookService.RemoveOpenOrder(m.BinanceService.OrderResponseToOrder(*order))
					m.OrderBookService.AddFilledOrder(m.BinanceService.OrderResponseToOrder(*order))
					err = m.WalletService.UpdateWallet()
					if err != nil {
						logger.Errorln("Error updating wallet" + err.Error())
						return
					}
					m.state.Current = MONITOR
					m.state.Time = int(time.Now().Unix())
					break
				}

				time.Sleep(1 * time.Second)
			}
		}
	}

}

func (m *MMStrategy) logAndList(msg string, loglevel log.Level) {

	switch loglevel {
	case log.PanicLevel:
		logger.Panicln(m.threadName + ": " + msg)
		break
	case log.FatalLevel:
		logger.Fatalln(m.threadName + ": " + msg)
		break
	case log.ErrorLevel:
		logger.Errorln(m.threadName + ": " + msg)
		break
	case log.WarnLevel:
		logger.Warnln(m.threadName + ": " + msg)
		break
	case log.InfoLevel:
		logger.Infoln(m.threadName + ": " + msg)
		m.logListMutex.Lock()
		*m.logList = append(*m.logList, m.threadName+": "+msg)
		m.logListMutex.Unlock()
		break
	case log.DebugLevel:
		logger.Debugln(m.threadName + ": " + msg)
		break
	case log.TraceLevel:
		logger.Traceln(m.threadName + ": " + msg)
		break
	}
}
