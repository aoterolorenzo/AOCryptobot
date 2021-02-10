package main

import (
	"./marketmaker"
	"./marketmaker/service/binance"
	"./marketmaker/service/common"
	"./marketmaker/ui"
	"os"
	"strings"
)

func main() {

	var logList []string
	binanceService := binance.BinanceService{}

	coin1 := strings.Split(os.Getenv("pair"), "-")[0]
	coin2 := strings.Split(os.Getenv("pair"), "-")[1]
	pair := strings.ReplaceAll(os.Getenv("pair"), "-", "")
	binanceService.SetPair(pair)
	binanceService.ConfigureClient()

	marketService := common.MarketService{}
	walletService := common.WalletService{Coin1: coin1, Coin2: coin2}
	walletService.InitWallet()
	_ = walletService.UpdateWallet()
	orderBookService := common.OrderBookService{}
	orderBookService.Init()

	marketService.StartMonitor(pair)
	for {
		if len(marketService.MarketSnapshotsRecord) > 0 {
			break
		}
	}

	go runBot(&binanceService, &marketService, &walletService, &orderBookService, &logList)
	runUI(&binanceService, &marketService, &walletService, &orderBookService, &logList)
}

func runBot(binanceService *binance.BinanceService, marketService *common.MarketService,
	walletService *common.WalletService, orderBookService *common.OrderBookService, logList *[]string) {
	strategy := marketmaker.MMStrategy{}
	strategy.SetServices(binanceService, marketService, walletService, orderBookService)
	strategy.SetLogList(logList)
	strategy.Execute()
}

func runUI(binanceService *binance.BinanceService, marketService *common.MarketService,
	walletService *common.WalletService, orderBookService *common.OrderBookService, logList *[]string) {
	userInterface := ui.UserInterface{}
	userInterface.SetServices(binanceService, marketService, walletService, orderBookService)
	userInterface.SetLogList(logList)
	userInterface.Run()
}
