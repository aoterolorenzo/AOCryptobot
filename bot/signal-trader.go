package bot

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	bot "gitlab.com/aoterocom/AOCryptobot/bot/services"
	"gitlab.com/aoterocom/AOCryptobot/database"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	binance2 "gitlab.com/aoterocom/AOCryptobot/providers/paper"
	"gitlab.com/aoterocom/AOCryptobot/services"
	strategies2 "gitlab.com/aoterocom/AOCryptobot/strategies"
	"os"
	"strconv"
	"strings"
)

type Bot struct {
}

func init() {

	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/conf.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
}

func (st *Bot) Run(c *cli.Context) {
	helpers.Logger.Infoln("🖖🏻 Signal Trader started")

	interval := os.Getenv("interval")
	strategiesString := c.String("strategies")
	if strategiesString == "" {
		strategiesString = os.Getenv("strategies")
	}

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

	var databaseService *database.DBService
	var err error
	databaseIsEnabled, _ := strconv.ParseBool(os.Getenv("enableDatabaseRecording"))

	if databaseIsEnabled == true {
		databaseService, err = database.NewDBService(os.Getenv("databaseHost"), os.Getenv("databasePort"), os.Getenv("databaseName"),
			os.Getenv("databaseUser"), os.Getenv("databasePassword"))
		if err != nil {
			panic(err)
		}
	}

	var strategies []interfaces.Strategy

	for _, strategy := range strategiesList {
		generatedStrategy, err := strategies2.StrategyFactory(strategy, interval)
		if err != nil {
			helpers.Logger.Errorln(err)
			log.Errorln(err)
			os.Exit(1)
		}
		strategies = append(strategies, generatedStrategy)
	}

	marketAnalysisService := services.NewMarketAnalysisService(exchangeService, strategies, &pairAnalysisResults, databaseService)
	marketAnalysisService.PopulateWithPairs(targetCoin, whitelistCoins, blacklistCoins)
	go marketAnalysisService.AnalyzeMarkets()

	mms := services.NewMultiMarketService(databaseService, &pairAnalysisResults, interval)
	go mms.StartMonitor()
	go mms.SignalAnalyzer()

	// trade on pairAnalysisResults
	trader := bot.NewBot(databaseService, &marketAnalysisService, &mms)
	trader.Start()
}
