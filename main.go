package main

import (
	"gitlab.com/aoterocom/AOCryptobot/bot_signaltrader"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
	"gitlab.com/aoterocom/AOCryptobot/services"
	strategies2 "gitlab.com/aoterocom/AOCryptobot/strategies"
)

func main() {
	//bot := &bot_oscillator.MarketMaker{}
	//rubBot(bot)
	helpers.Logger.Infoln("🖖🏻 Bot started")

	pairAnalysisResults := []*analytics.PairAnalysis{}
	bs := binance.NewBinanceService()
	exchangeService := interfaces.ExchangeService(&bs)
	exchangeService.ConfigureClient()
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
	marketAnalysisService.PopulateWithPairs("EUR")
	go marketAnalysisService.AnalyzeMarkets()

	mms := services.NewMultiMarketService(&pairAnalysisResults)
	go mms.StartMonitor()

	// trade on pairAnalysisResults
	trader := bot_signaltrader.NewTrader(&marketAnalysisService, &mms)
	trader.Start()
}

func rubBot(b interfaces.MarketBot) {
	b.Run()
}
