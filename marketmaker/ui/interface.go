package ui

import (
	"../model"
	"../service/binance"
	"../service/common"
	"fmt"
	"github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"log"
	"time"
)

type UserInterface struct {
	ExchangeService  *binance.BinanceService
	MarketService    *common.MarketService
	WalletService    *common.WalletService
	OrderBookService *common.OrderBookService
	initialWallet    *model.Wallet
	initialPrice     float64
	logList          *[]string
	initialBalance   float64
	currentBalance   float64
	totalPyL         float64
	totalPyLPct      float64
}

func (ui *UserInterface) SetExchangeService(exchangeService *binance.BinanceService) {
	ui.ExchangeService = exchangeService
}

func (ui *UserInterface) SetServices(exchangeService *binance.BinanceService, MarketService *common.MarketService,
	walletService *common.WalletService, orderBookService *common.OrderBookService) {
	ui.ExchangeService = exchangeService
	ui.MarketService = MarketService
	ui.WalletService = walletService
	ui.OrderBookService = orderBookService
	ui.initialWallet = ui.WalletService.Wallet
	ui.initialBalance = ui.WalletService.GetTotalAssetsBalance(ui.MarketService.CurrentPrice(&ui.MarketService.MarketSnapshotsRecord))

	ui.UpdatePyL()
}

func (ui *UserInterface) UpdatePyL() {

	ui.currentBalance = ui.WalletService.GetTotalAssetsBalance(ui.MarketService.CurrentPrice(&ui.MarketService.MarketSnapshotsRecord))
	ui.totalPyL = ui.currentBalance - ui.initialBalance
	ui.totalPyLPct = (ui.currentBalance * 100 / ui.initialBalance) - 100.0
}

func (ui *UserInterface) SetLogList(logList *[]string) {
	ui.logList = logList
}
func (ui *UserInterface) Run() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	ui.ExchangeService.ConfigureClient()

	for {
		time.Sleep(1 * time.Second)

		maxPrice, err := ui.MarketService.MaxPrice(60, &ui.MarketService.MarketSnapshotsRecord)
		if err != nil {
			log.Fatalln(err)
		}
		minPrice, err := ui.MarketService.MinPrice(60, &ui.MarketService.MarketSnapshotsRecord)
		if err != nil {
			log.Fatalln(err)
		}
		oscilation, err := ui.MarketService.PctVariation(60, &ui.MarketService.MarketSnapshotsRecord)
		if err != nil {
			log.Fatalln(err)
		}
		lowerAsk := ui.MarketService.MarketSnapshotsRecord[0].LowerAskPrice
		centerPrice := ui.MarketService.MarketSnapshotsRecord[0].CenterPrice
		higherBid := ui.MarketService.MarketSnapshotsRecord[0].HigherBidPrice
		percentil, err := ui.MarketService.CurrentPricePercentile(60, &ui.MarketService.MarketSnapshotsRecord)
		if err != nil {
			log.Fatalln(err)
		}

		marketSnapshotParagraph := widgets.NewParagraph()
		marketSnapshotParagraph.BorderStyle.Fg = termui.ColorYellow
		marketSnapshotParagraph.TitleStyle.Fg = termui.ColorYellow
		marketSnapshotParagraph.Block.Title = "Market Status " + ui.WalletService.Coin1 + ui.WalletService.Coin2
		marketSnapshotParagraph.Text = fmt.Sprintf("Max Price: %.8f\n", maxPrice)
		marketSnapshotParagraph.Text += fmt.Sprintf("Min Price: %.8f\n", minPrice)
		marketSnapshotParagraph.Text += fmt.Sprintf("Oscilation: %.2f%%\n", oscilation)
		marketSnapshotParagraph.Text += fmt.Sprintf("Lower Ask: %.8f\n", lowerAsk)
		marketSnapshotParagraph.Text += fmt.Sprintf("[Current Price: %.8f](fg:blue)\n", centerPrice)
		marketSnapshotParagraph.Text += fmt.Sprintf("Higher Bid: %.8f\n", higherBid)
		marketSnapshotParagraph.Text += fmt.Sprintf("Percentil: %2f\n", percentil)
		marketSnapshotParagraph.SetRect(0, 0, 34, 9)

		balanceCoin1, err := ui.WalletService.GetAssetBalance(ui.WalletService.Coin1)
		if err != nil {
			log.Fatalln(err)
		}
		balanceCoin2, err := ui.WalletService.GetAssetBalance(ui.WalletService.Coin2)
		if err != nil {
			log.Fatalln(err)
		}

		walletStatusParagraph := widgets.NewParagraph()
		walletStatusParagraph.Block.Title = "Wallet"
		walletStatusParagraph.Text = fmt.Sprintf("Balance %s: %.8f\n", ui.WalletService.Coin1, balanceCoin1)
		walletStatusParagraph.Text += fmt.Sprintf("Balance %s: %.8f\n", ui.WalletService.Coin2, balanceCoin2)
		walletStatusParagraph.SetRect(68, 0, 34, 4)

		operationsList := widgets.NewList()
		operationsList.Block.Title = "Operations"
		operationsList.Rows = *ui.logList
		operationsList.SetRect(0, 18, 100, 9)

		ordersParagraph := widgets.NewParagraph()
		ordersParagraph.Block.Title = "Open Orders"
		ordersParagraph.Text += fmt.Sprintf("Total: %d\n", ui.OrderBookService.OpenOrdersCount())
		ordersParagraph.Text += fmt.Sprintf("Sells: %d\n", ui.OrderBookService.OpenSellOrdersCount())
		ordersParagraph.Text += fmt.Sprintf("Buys: %d", ui.OrderBookService.OpenBuyOrdersCount())
		ordersParagraph.SetRect(68, 9, 34, 4)

		if time.Now().Second()%10 == 0 {
			ui.UpdatePyL()
		}

		pAndLParagraph := widgets.NewParagraph()
		pAndLParagraph.Block.Title = "P&L"
		pAndLParagraph.Text = fmt.Sprintf("Initial: %.8f %s\n",
			ui.initialBalance, ui.WalletService.Coin2)
		pAndLParagraph.Text += fmt.Sprintf("Current: %.8f %s\n",
			ui.currentBalance, ui.WalletService.Coin2)
		pAndLParagraph.Text += fmt.Sprintf("Total P&L: %2f€\n", ui.totalPyL)
		pAndLParagraph.Text += fmt.Sprintf("%% P&L: %.2f%%\n", ui.totalPyLPct)
		pAndLParagraph.SetRect(100, 6, 68, 0)

		operationsList.ScrollBottom()

		termui.Render(marketSnapshotParagraph, walletStatusParagraph, operationsList, ordersParagraph, pAndLParagraph)

	}
}
