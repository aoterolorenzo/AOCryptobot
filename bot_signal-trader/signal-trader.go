package bot_signal_trader

import (
	"github.com/joho/godotenv"
	signalTrader "gitlab.com/aoterocom/AOCryptobot/bot_signal-trader/services"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	binance2 "gitlab.com/aoterocom/AOCryptobot/providers/paper"
	"gitlab.com/aoterocom/AOCryptobot/services"
	strategies2 "gitlab.com/aoterocom/AOCryptobot/strategies"
	"log"
	"os"
)

type SignalTrader struct {
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/bot_signal-trader/conf.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
}

func (st *SignalTrader) Run() {
	//bot := &bot_oscillator.MarketMaker{}
	//rubBot(bot)
	helpers.Logger.Infoln("🖖🏻 Signal Trader started")

	targetCoin := os.Getenv("targetCoin")
	var pairAnalysisResults []*analytics.PairAnalysis
	bs := binance2.NewPaperService()
	exchangeService := interfaces.ExchangeService(bs)
	lun1MarCustomStrategy := strategies2.NewLun1MarCustomStrategy()
	MACDCustomStrategy := strategies2.NewMACDCustomStrategy()
	stableStrategy := strategies2.NewStableStrategy()
	stochRSICustomStrategy := strategies2.NewStochRSICustomStrategy()

	strategies := []interfaces.Strategy{
		&lun1MarCustomStrategy,
		&stochRSICustomStrategy,
		&MACDCustomStrategy,
		&stableStrategy,
	}
	marketAnalysisService := services.NewMarketAnalysisService(exchangeService, strategies, &pairAnalysisResults)
	marketAnalysisService.PopulateWithPairs(targetCoin)
	go marketAnalysisService.AnalyzeMarkets()

	mms := services.NewMultiMarketService(&pairAnalysisResults)
	go mms.StartMonitor()

	// trade on pairAnalysisResults
	trader := signalTrader.NewSignalTrader(&marketAnalysisService, &mms)
	trader.Start()
}
