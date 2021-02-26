package main

import (
	"gitlab.com/aoterocom/AOCryptobot/bot_signaltrader"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
	"gitlab.com/aoterocom/AOCryptobot/services"
	strategies2 "gitlab.com/aoterocom/AOCryptobot/strategies"
)

func main() {
	//bot := &bot_oscillator.MarketMaker{}
	//rubBot(bot)

	pairAnalysisResults := []*analytics.PairAnalysis{}
	exchangeService := interfaces.ExchangeService(&binance.BinanceService{})
	exchangeService.ConfigureClient()
	strategies := []interfaces.Strategy{
		&strategies2.StochRSICustomStrategy{},
		&strategies2.MACDCustomStrategy{},
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
