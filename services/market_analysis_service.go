package services

import (
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
)

type MarketAnalysisService struct {
	PairStrategyAnalysisList *[]analytics.PairAnalysisResult
	exchangeService          interfaces.ExchangeService
	strategies               []interfaces.Strategy
}

func (mas MarketAnalysisService) NewMarketAnalysisService(exchangeService interfaces.ExchangeService,
	strategies []interfaces.Strategy) MarketAnalysisService {
	mas.exchangeService = exchangeService
	mas.strategies = strategies
	return mas
}

func (mas *MarketAnalysisService) AnalyzePair(pair string) {

}

func (mas *MarketAnalysisService) PopulateWithPairs(coin string) {
	for _, pair := range mas.exchangeService.GetMarkets(coin) {
		pairAnalysis := analytics.PairAnalysisResult{Pair: pair}
		pairAnalysis.TradeSignal = false
		*mas.PairStrategyAnalysisList = append(*mas.PairStrategyAnalysisList, pairAnalysis)
	}
}

func (mas *MarketAnalysisService) AnalyzeMarkets() {
	for _, pairAnalysis := range *mas.PairStrategyAnalysisList {
		mas.AnalyzePair(pairAnalysis.Pair)
	}
}
