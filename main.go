package main

import (
	"./marketmaker"
	"./marketmaker/service/binance"
	"./marketmaker/service/common"
	"./marketmaker/ui"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {

	log.Println("Loading bot...")
	threadNumber, err := strconv.Atoi(os.Getenv("threadNumber"))
	monitorWindow, err := strconv.Atoi(os.Getenv("monitorWindow"))
	if err != nil {
		log.Fatal("Error parsing threadNumber from strategy.env")
	}
	var logList []string
	logListMutex := sync.Mutex{}
	orderBookMutex := sync.Mutex{}
	binanceService := binance.BinanceService{}

	coin1 := strings.Split(os.Getenv("pair"), "-")[0]
	coin2 := strings.Split(os.Getenv("pair"), "-")[1]
	pair := strings.ReplaceAll(os.Getenv("pair"), "-", "")
	binanceService.SetPair(pair)
	binanceService.ConfigureClient()

	marketService := common.MarketService{}
	walletService := common.WalletService{Coin1: coin1, Coin2: coin2}
	walletService.InitWallet()
	err = walletService.UpdateWallet()
	if err != nil {
		log.Fatal("Error initially updating wallet")
	}
	orderBookService := common.OrderBookService{}
	orderBookService.SetMutex(&orderBookMutex)
	orderBookService.Init()

	marketService.StartMonitor(pair)
	for {
		time.Sleep(1 * time.Second)
		if len(marketService.MarketSnapshotsRecord) > 0 {
			break
		}
	}

	for i := 1; i <= threadNumber; i++ {
		threadName := "Thread " + strconv.Itoa(i)
		waitTime := (monitorWindow / threadNumber * i) - monitorWindow/threadNumber
		go runBot(&binanceService, &marketService, &walletService, &orderBookService, &logList,
			threadName, &logListMutex, waitTime)
	}

	runUI(&binanceService, &marketService, &walletService, &orderBookService, &logList)
}

func runBot(binanceService *binance.BinanceService, marketService *common.MarketService,
	walletService *common.WalletService, orderBookService *common.OrderBookService,
	logList *[]string, threadName string, logListMutex *sync.Mutex, waitTime int) {
	strategy := marketmaker.MMStrategy{}
	strategy.SetServices(binanceService, marketService, walletService, orderBookService)
	strategy.SetLogListAndMutex(logList, logListMutex)
	strategy.SetThreadName(threadName)
	strategy.Execute(waitTime)
}

func runUI(binanceService *binance.BinanceService, marketService *common.MarketService,
	walletService *common.WalletService, orderBookService *common.OrderBookService, logList *[]string) {
	userInterface := ui.UserInterface{}
	userInterface.SetServices(binanceService, marketService, walletService, orderBookService)
	userInterface.SetLogList(logList)
	userInterface.Run()
}
