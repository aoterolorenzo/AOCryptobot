package bot_oscillator

import (
	"github.com/sdcoffey/techan"
	marketMakerServices "gitlab.com/aoterocom/AOCryptobot/bot_oscillator/services"
	"gitlab.com/aoterocom/AOCryptobot/bot_oscillator/ui"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
	"gitlab.com/aoterocom/AOCryptobot/services"
	"gitlab.com/aoterocom/AOCryptobot/strategies"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MarketMaker struct {
	exchangeProvider interfaces.ExchangeService
	marketService    *services.SingleMarketService
	walletService    *services.WalletService
	orderBookService *services.OrderBookService
	logList          []string
	threadName       string
	logListMutex     *sync.Mutex
	waitTime         int
}

func (mm *MarketMaker) Run() {

	threadNumber, err := strconv.Atoi(os.Getenv("threadNumber"))
	monitorWindow, err := strconv.Atoi(os.Getenv("monitorWindow"))
	if err != nil {
		helpers.Logger.Fatalln("Error parsing threadNumber from strategy.env")
	}

	logListMutex := sync.Mutex{}
	orderBookMutex := sync.Mutex{}

	bs := binance.NewBinanceService()
	mm.exchangeProvider = bs

	coin1 := strings.Split(os.Getenv("pair"), "-")[0]
	coin2 := strings.Split(os.Getenv("pair"), "-")[1]
	pair := strings.ReplaceAll(os.Getenv("pair"), "-", "")

	sms := services.NewSingleMarketService(*techan.NewTimeSeries(), pair)
	mm.marketService = &sms
	ws := services.NewWalletService(coin1, coin2)
	mm.walletService = &ws
	mm.walletService.InitWallet()
	err = mm.walletService.UpdateWallet()
	if err != nil {
		helpers.Logger.Fatalln("Error initially updating wallet" + err.Error())
	}
	mm.orderBookService = &services.OrderBookService{}
	mm.orderBookService.SetMutex(&orderBookMutex)
	mm.orderBookService.Init()

	mm.marketService.StartMultiMonitor(pair)
	for {
		time.Sleep(1 * time.Second)
		if len(mm.marketService.MarketSnapshotsRecord) > 0 {
			break
		}
	}

	for i := 1; i <= threadNumber; i++ {
		threadName := "Thread " + strconv.Itoa(i)
		waitTime := (monitorWindow / threadNumber * i) - monitorWindow/threadNumber
		go mm.runThread(threadName, &logListMutex, waitTime)
	}

	mm.runUI(&mm.logList)
}

func (mm *MarketMaker) runThread(threadName string, logListMutex *sync.Mutex, waitTime int) {
	marketMakerService := marketMakerServices.MarketMakerService{}
	marketMakerService.SetServices(mm.exchangeProvider, mm.marketService, mm.walletService, mm.orderBookService)
	marketMakerService.SetStrategy(&strategies.StochRSICustomStrategy{})
	marketMakerService.SetLogListAndMutex(&mm.logList, logListMutex)
	marketMakerService.SetThreadName(threadName)
	marketMakerService.Execute(waitTime)
}

func (mm *MarketMaker) runUI(logList *[]string) {
	userInterface := ui.UserInterface{}
	userInterface.SetServices(mm.exchangeProvider, mm.marketService, mm.walletService, mm.orderBookService)
	userInterface.SetLogList(logList)
	userInterface.Run()
}
