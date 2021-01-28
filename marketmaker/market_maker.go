package marketmaker

import (
	"../marketmaker/helpers"
	service "../marketmaker/service/binance"
	"fmt"
	"github.com/adshao/go-binance/v2"
	tm "github.com/buger/goterm"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)


const(
	MONITOR = "MONITOR"
	BUYING = "BUYING"
	HOLDING = "HOLDING"
)

const(
	FILLED = "FILLED"
	CANCELED = "CANCELED"
	NEXT = "NEXT"
)

type MMStrategy struct {
	MarketService       *service.MarketService
	state               helpers.STATE
	startEURAmout       float64
	monitorWindow       int
	pctAmountToTrade    float64
	buyMargin           float64
	sellMargin          float64
	stopLossPct         float64
	trailingStopLossPct float64
	topPercentileToTrade   float64
	lowPercentileToTrade   float64
	panicModeMargin     float64
	monitorFrequency    int
	buyingTimeout       int
	sellingTimeout 		int

	// Execution time variables
	buyRate 			float64
	buyAmount 			float64
	buyOrderId 			int64

	sellRate       float64
	sellAmount     float64
	sellOCOOrder   *binance.CreateOCOResponse
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

func (m *MMStrategy) SetMarketService(marketService *service.MarketService)  {
	m.MarketService = marketService
}
func (m *MMStrategy) Execute()  {

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
	m.startEURAmout, _ = m.MarketService.GetTotalBalance("EUR")
	m.startEURAmout *= 1 - 0.016


	// Set initial state to MONITOR and start MarketService monitor
	m.state.Current = MONITOR


	// Wait window iterations until monitor loads
	fmt.Println("Loading window data...")

	for {
		if len(m.MarketService.MarketStatusList) > m.monitorWindow {
			break
		}
	}

	fmt.Println("Data loaded, bot started")

	for {
		// Switch functions depending on the current state
		tm.Clear()
		m.printInfo()
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

func (m *MMStrategy) monitor()  {

	lastPricePercentile, _ := m.MarketService.Analytics.CurrentPricePercentile(m.monitorWindow, &m.MarketService.MarketStatusList)
	pctVariation, _ := m.MarketService.Analytics.PctVariation(m.monitorWindow, &m.MarketService.MarketStatusList)

	if lastPricePercentile < m.topPercentileToTrade && lastPricePercentile > m.topPercentileToTrade {
		fmt.Printf("Mercado fuera de percentil recomendado para iniciar compra %.2f\n", lastPricePercentile)
		return
	}

	if pctVariation < m.panicModeMargin {
		println("Mercado en caída!")
		return
	}

	fmt.Println("Momento de comprar!!!!!")


	m.buyRate = m.MarketService.Analytics.CurrentPrice(&m.MarketService.MarketStatusList) * (1 - m.buyMargin)
	balanceA, _ := m.MarketService.GetTotalBalance("EUR")
	m.buyAmount =  balanceA * m.pctAmountToTrade / 100

	if m.startEURAmout > balanceA{
		fmt.Println("SALIENDO!!! DETECTADAS PERDIDAS!!")
		os.Exit(1)
	}

	buyOrderId, err := m.MarketService.MakeOrder(m.buyAmount, m.buyRate, binance.SideTypeBuy)
	if err != nil {
		println(1)
		fmt.Println(err)
		return
	}
	m.buyOrderId = buyOrderId

	fmt.Printf("Emitida orden de compra por %f EUR, cantidad %f EUR\n", m.buyRate, m.buyAmount)
	fmt.Println("Esperando compra...")

	m.state.Time = int(time.Now().Unix())
	m.state.Current = BUYING

}

func (m *MMStrategy) buying()  {

	orderStatus, err := m.MarketService.GetOrderStatus(m.buyOrderId)
	if err != nil {
		println(2)
		fmt.Println(err)
	}

	if orderStatus == "FILLED" {

		fmt.Print("Precio alcanzado. Compra realizada.")
		balanceA, err := m.MarketService.GetTotalBalance("ETH")
		if err != nil {
			println(3)
			fmt.Println(err)
		}
		balanceB, err := m.MarketService.GetTotalBalance("EUR")
		if err != nil {
			println(4)
			fmt.Println(err)
		}

		fmt.Printf("Saldo: %f ETH , %f EUR", balanceA, balanceB)
		fmt.Printf("\nPrecio de compra: %f EUR", m.buyRate)
		fmt.Println()

		//INICIAMOS ORDEN DE VENTA

		m.sellRate = m.buyRate * (1 + m.buyMargin) * (1 + m.sellMargin)
		m.sellAmount = m.buyAmount / m.buyRate
		m.stopPrice = m.sellRate * (1 - m.stopLossPct)
		m.stopLimitPrice = m.stopPrice * (1 + m.trailingStopLossPct) * (1-0.0004)

		m.sellOCOOrder, err = m.MarketService.MakeOCOOrder(m.sellAmount, m.sellRate, m.stopPrice, m.stopLimitPrice, binance.SideTypeSell)
		if err != nil {
			println(5)
			fmt.Println(err)
		}

		fmt.Printf("Emitiendo orden de venta OCO por  %f EUR, cantidad %f EUR, stop-loss %f EUR",
			m.sellRate, m.sellAmount, m.stopPrice)
		fmt.Println("\nEsperando venta...")


		m.state.Current = HOLDING
		m.state.Time = int(time.Now().Unix())

	} else if m.state.Time + m.buyingTimeout < int(time.Now().Unix()) {

		fmt.Println("Buy timeout. Canceling...")
		err = m.MarketService.CancelOrder(m.buyOrderId)
		if err != nil {
			println(55)
			fmt.Println(err)
			return
		}

		m.state.Current = MONITOR
		m.state.Time = int(time.Now().Unix())
	}

}

func (m *MMStrategy) holding()  {


	for _, order := range m.sellOCOOrder.Orders {
		// GET FIRST ORDER
		tempSellOrder, err := m.MarketService.GetOrder(order.OrderID)
		if err != nil {
			println(6)
			fmt.Println(err)
			return
		}

		if tempSellOrder.Status == binance.OrderStatusTypeFilled &&
			tempSellOrder.Type == binance.OrderTypeStopLossLimit {
			// STOP LOSS LIMIT FILLED
			fmt.Println("STOP LOSS: Sold by strategy")
			m.state.Current = MONITOR
			m.state.Time = int(time.Now().Unix())
		} else if tempSellOrder.Status == binance.OrderStatusTypeFilled {
			// LIMIT FILLED
			fmt.Println("TAKING BENEFIT: Sell ends successfully")
			m.state.Current = MONITOR
			m.state.Time = int(time.Now().Unix())
		}

	}

	// CHECK SELL IS NOT TIMEOUT
	/*if m.sellingTimeout == 0 || m.state.Time + m.sellingTimeout > int(time.Now().Unix()){


	} else if false{
		fmt.Println("Sell timeout.")

		orderA, err := m.MarketService.GetOrder(m.sellOCOOrder.Orders[0].OrderID)
		if err != nil {
			println(7)
			fmt.Println(err)
		}

		orderB, err := m.MarketService.GetOrder(m.sellOCOOrder.Orders[0].OrderID)
		if err != nil {
			println(8)
			fmt.Println(err)
		}

		if !(orderA.Status == binance.OrderStatusTypePartiallyFilled) &&
			!(orderB.Status == binance.OrderStatusTypePartiallyFilled){
			_ = m.MarketService.CancelOrder(orderA.OrderID)
			_ = m.MarketService.CancelOrder(orderB.OrderID)

			fmt.Println("Sell orders have been canceled")
			fmt.Print("Selling at market price ")

			orderId, err := m.MarketService.MakeOrder(m.sellAmount,
				m.MarketService.MarketStatusList[0].HigherBidPrice, binance.SideTypeSell)
			if err != nil {
				println(11)
				fmt.Println(err)
			}

			fmt.Print("and waiting to fill...")
			for {
				timeoutSellORder , err := m.MarketService.GetOrder(orderId)
				if err != nil {
					println(12)
					fmt.Println(err)
					break
				}

				if timeoutSellORder.Status == binance.OrderStatusTypeFilled {
					fmt.Print("Timeout sell order filled")
					m.state.Current = MONITOR
					m.state.Time = int(time.Now().Unix())
					break
				}

				time.Sleep(1 * time.Second)
			}
		}
	}*/

}

func (m MMStrategy) printInfo() {
	fmt.Println("####################")
	fmt.Println("Estado del mercado:")
	fmt.Println("####################")

	fmt.Print("El precio máximo alcanzado en los últimos 5 minutos es:")
	fmt.Println(m.MarketService.Analytics.MaxPrice(m.monitorWindow, &m.MarketService.MarketStatusList))

	fmt.Print("El precio mínimo alcanzado en los últimos 5 minutos es:")
	fmt.Println(m.MarketService.Analytics.MinPrice(m.monitorWindow, &m.MarketService.MarketStatusList))

	fmt.Print("El mercado ha variado en un :")
	fmt.Print( m.MarketService.Analytics.PctVariation(m.monitorWindow, &m.MarketService.MarketStatusList))
	fmt.Println(" %")


	fmt.Print("LowerAskPrice:")
	fmt.Println(m.MarketService.MarketStatusList[0].LowerAskPrice)
	fmt.Print("Precio actual:")
	fmt.Println(m.MarketService.MarketStatusList[0].CenterPrice)
	fmt.Print("HigherBidPrice:")
	fmt.Println(m.MarketService.MarketStatusList[0].HigherBidPrice)
	fmt.Print("Spread:")
	fmt.Println(m.MarketService.MarketStatusList[0].Spread)
	fmt.Print("El precio actual está en el percentil:")
	fmt.Print( m.MarketService.Analytics.CurrentPricePercentile(m.monitorWindow, &m.MarketService.MarketStatusList))
	fmt.Println(" de las variaciones del período de monitorización")
	fmt.Print("Saldo EUR:")
	fmt.Println(m.MarketService.GetTotalBalance("EUR"))
	fmt.Print("Saldo ETH:")
	fmt.Println(m.MarketService.GetTotalBalance("ETH"))
}



