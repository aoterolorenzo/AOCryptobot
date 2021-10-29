package bot_signal_trader

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	signalTrader "gitlab.com/aoterocom/AOCryptobot/bot_signal-trader/services"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	binance2 "gitlab.com/aoterocom/AOCryptobot/providers/paper"
	"gitlab.com/aoterocom/AOCryptobot/services"
	strategies2 "gitlab.com/aoterocom/AOCryptobot/strategies"
	"os"
	"strings"
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
	helpers.Logger.Infoln("🖖🏻 Signal Trader started")

	strategiesString := os.Getenv("strategies")
	strategiesList := strings.Split(strategiesString, ",")
	if strategiesList[0] == "" {
		helpers.Logger.Infoln("error: couldn't initialize bot. No strategies set")
	}

	targetCoin := os.Getenv("targetCoin")
	whitelistCoinsString := os.Getenv("whitelistCoins")
	blacklistCoinsString := os.Getenv("blacklistCoins")
	whitelistCoins := strings.Split(whitelistCoinsString, ",")
	if whitelistCoins[0] == "" {
		whitelistCoins = []string{}
	}

	blacklistCoins := strings.Split(blacklistCoinsString, ",")
	if blacklistCoins[0] == "" {
		blacklistCoins = []string{}
	}

	var pairAnalysisResults []*analytics.PairAnalysis
	bs := binance2.NewPaperService()
	exchangeService := interfaces.ExchangeService(bs)

	var strategies []interfaces.Strategy

	for _, strategy := range strategiesList {
		generatedStrategy, err := strategies2.StrategyFactory(strategy)
		if err != nil {
			helpers.Logger.Errorln(err)
			log.Errorln(err)
			os.Exit(1)
		}
		strategies = append(strategies, generatedStrategy)
	}

	marketAnalysisService := services.NewMarketAnalysisService(exchangeService, strategies, &pairAnalysisResults)
	marketAnalysisService.PopulateWithPairs(targetCoin, whitelistCoins, blacklistCoins)
	go marketAnalysisService.AnalyzeMarkets()

	mms := services.NewMultiMarketService(&pairAnalysisResults)
	go mms.StartMonitor()

	// trade on pairAnalysisResults
	trader := signalTrader.NewSignalTrader(&marketAnalysisService, &mms)
	trader.Start()
}
