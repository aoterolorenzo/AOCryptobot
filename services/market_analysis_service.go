package services

import (
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"sort"
)

type MarketAnalysisService struct {
	PairAnalysisResults *[]*analytics.PairAnalysis
	exchangeService     interfaces.ExchangeService
	strategies          []interfaces.Strategy
}

func NewMarketAnalysisService(exchangeService interfaces.ExchangeService,
	strategies []interfaces.Strategy, pairAnalysisResults *[]*analytics.PairAnalysis) MarketAnalysisService {
	return MarketAnalysisService{
		PairAnalysisResults: pairAnalysisResults,
		exchangeService:     exchangeService,
		strategies:          strategies,
	}
}

func (mas *MarketAnalysisService) PopulateWithPairs(coin string) {
	for _, pair := range mas.exchangeService.GetMarkets(coin) {
		pairAnalysis := analytics.PairAnalysis{Pair: pair}
		pairAnalysis.TradeSignal = false
		*mas.PairAnalysisResults = append(*mas.PairAnalysisResults, &pairAnalysis)
	}
}

func (mas *MarketAnalysisService) AnalyzeMarkets() {
	for {
		for _, pairAnalysis := range *mas.PairAnalysisResults {
			pairAnalysisResultPtr := mas.getPairAnalysisResult(pairAnalysis.Pair)
			newPairAnalysisResult, err := mas.analyzePair(pairAnalysis.Pair)
			if err == nil {
				*pairAnalysisResultPtr = newPairAnalysisResult
			}
		}
	}

	//D for _ , item := range mas.GetTradeSignaledMarketsByInvStdDev() {
	//D 	fmt.Printf("Best strategy for %s: %s Tradeable: %t.\n", item.Pair, item.BestStrategy,
	//D 		item.TradeSignal)
	//D }
}

func (mas *MarketAnalysisService) analyzePair(pair string) (analytics.PairAnalysis, error) {
	pairAnalysisResult := analytics.PairAnalysis{
		StrategiesAnalysis: nil,
		TradeSignal:        false,
		Pair:               pair,
	}

	// For each strategy in pair
	for _, strategy := range mas.strategies {
		//We analyze the strategy and set the results in pairAnalysisResult
		mas.exchangeService.SetPair(pair)
		helpers.Logger.Debugln(pair + " Market")
		strategyAnalysisResult, err := strategy.Analyze(mas.exchangeService)
		if err != nil {
			helpers.Logger.Errorln(err.Error())
			return pairAnalysisResult, err
		}
		pairAnalysisResult.StrategiesAnalysis = append(pairAnalysisResult.StrategiesAnalysis,
			*strategyAnalysisResult)
	}

	//Once the results are set for the pair, we check the strategies and choose between them if they are valid
	chosenStrategy := mas.chooseStrategy(pairAnalysisResult)
	if chosenStrategy != nil {
		pairAnalysisResult.TradeSignal = true
		pairAnalysisResult.BestStrategy = chosenStrategy
	}

	return pairAnalysisResult, nil
}

func (mas *MarketAnalysisService) chooseStrategy(pairAnalysisResult analytics.PairAnalysis) interface{} {
	var betterStrategy interface{}
	betterStrategy = nil
	betterMeanStdDevRelation := -10000.0
	for _, strategy := range pairAnalysisResult.StrategiesAnalysis {
		//D fmt.Printf("Strategy: %s Ratio: %.2f (Analyze %t)\n",strategy.Strategy, strategy.Mean / strategy.StdDev,  strategy.Analyze)
		if strategy.IsCandidate && strategy.Mean/strategy.StdDev > betterMeanStdDevRelation {
			//D fmt.Printf("Strategy %s with ratio %.2f%% is better than %.2f\n", strategy.Strategy,
			//D 	strategy.Mean / strategy.StdDev, betterMeanStdDevRelation)
			betterMeanStdDevRelation = strategy.Mean / strategy.StdDev

			betterStrategy = strategy.Strategy
		} else {
			//D fmt.Printf("Strategy %s with ratio %.2f%% is worst than %.2f\n", strategy.Strategy,
			//D 	strategy.Mean / strategy.StdDev, betterMeanStdDevRelation)
		}
	}
	//D fmt.Printf("Better Strategy: %s\n", betterStrategy)

	return betterStrategy
}
func positiveNegativeRatio(list []float64) float64 {
	countPositive := 0
	countNegative := 0
	for _, item := range list {
		if item > 0 {
			countPositive++
		} else {
			countNegative++
		}
	}

	if countNegative == 0 {
		return 0
	}
	return float64(countPositive) / float64(countNegative)
}
func (mas *MarketAnalysisService) getPairAnalysisResult(pair string) *analytics.PairAnalysis {
	for _, pairStrategyAnalysis := range *mas.PairAnalysisResults {
		if pairStrategyAnalysis.Pair == pair {
			return pairStrategyAnalysis
		}
	}
	return nil
}

func (mas *MarketAnalysisService) getStrategyAnalysis(pairAnalysisResult analytics.PairAnalysis, strategy interfaces.Strategy) *analytics.StrategyAnalysis {
	for _, strategyAnalysisResult := range pairAnalysisResult.StrategiesAnalysis {
		if strategyAnalysisResult.Strategy == strategy {
			return &strategyAnalysisResult
		}
	}
	return nil
}
func (mas *MarketAnalysisService) GetBestStrategyResults(pairAnalysisResult analytics.PairAnalysis) analytics.StrategyAnalysis {
	for _, strategyResults := range pairAnalysisResult.StrategiesAnalysis {
		if pairAnalysisResult.BestStrategy == strategyResults.Strategy {
			return strategyResults
		}
	}
	return analytics.StrategyAnalysis{}
}

func (mas *MarketAnalysisService) GetTradeSignaledMarketsByInvStdDev() []analytics.PairAnalysis {

	pairAnalysisResults := []analytics.PairAnalysis{}
	for _, result := range *mas.PairAnalysisResults {
		if result.TradeSignal {
			pairAnalysisResults = append(pairAnalysisResults, *result)
		}
	}

	sort.Slice(pairAnalysisResults, func(i, j int) bool {

		bestStrategy1Results := mas.GetBestStrategyResults(pairAnalysisResults[i])
		bestStrategy2Results := mas.GetBestStrategyResults(pairAnalysisResults[j])

		return bestStrategy1Results.StdDev > bestStrategy2Results.StdDev

	})

	return pairAnalysisResults

}
