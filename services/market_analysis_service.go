package services

import (
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"math"
	"sort"
)

var logger = helpers.Logger{}

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
		//D fmt.Println(pair + ":")
		strategyAnalysisResult, err := mas.analyzeStrategy(strategy)
		if err != nil {
			logger.Errorln(err.Error())
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
		//D fmt.Printf("Strategy: %s Ratio: %.2f (IsCandidate %t)\n",strategy.Strategy, strategy.Mean / strategy.StdDev,  strategy.IsCandidate)
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

func (mas *MarketAnalysisService) analyzeStrategy(strategy interfaces.Strategy) (*analytics.StrategyAnalysis, error) {

	strategyAnalysisResults := analytics.StrategyAnalysis{
		IsCandidate: false,
		Strategy:    strategy,
	}

	//D Printf("Analyzing %s\n", strategyAnalysisResults.Strategy)

	result15m1000, err := strategy.PerformAnalysis(mas.exchangeService, "15m", 1000, 0, nil)
	if err != nil {
		return nil, err
	}
	debugPrint(result15m1000, result15m1000.Profit)
	strategyAnalysisResults.StrategyResults = append(strategyAnalysisResults.StrategyResults, result15m1000)

	result15m500, err := strategy.PerformAnalysis(mas.exchangeService, "15m", 500, 0, &result15m1000.Constants)
	if err != nil {
		return nil, err
	}
	debugPrint(result15m500, result15m500.Profit)
	strategyAnalysisResults.StrategyResults = append(strategyAnalysisResults.StrategyResults, result15m500)

	data := []float64{
		result15m1000.Profit,
		result15m500.Profit,
	}
	sum := sum(data)
	strategyAnalysisResults.Mean = sum / float64(len(data))
	strategyAnalysisResults.StdDev = stdDev(data, strategyAnalysisResults.Mean)
	//D fmt.Printf("3 Mean: %f Deviation %f Ratio %f\n", strategyAnalysisResults.Mean,
	//D strategyAnalysisResults.StdDev, strategyAnalysisResults.Mean / strategyAnalysisResults.StdDev)

	// Check if strategy pass conditions
	// If all profits are positive
	if result15m1000.Profit > 0.0 && result15m500.Profit > 0.0 &&
		strategyAnalysisResults.Mean/0.6 > strategyAnalysisResults.StdDev &&
		positiveNegativeRatio(result15m1000.ProfitList) >= 1.0 &&
		positiveNegativeRatio(result15m500.ProfitList) >= 1.0 {
		//mean > 25.0 && strategyResults150.Profit > 12.0 || {
		strategyAnalysisResults.IsCandidate = true
		//fmt.Printf("!candidata prof1000 %f prof500 %f mean/0.6 %f stddev %f ratio1000 %f ratio500 %f %v\n", result15m1000.Profit, result15m500.Profit,
		//	strategyAnalysisResults.Mean/0.6, strategyAnalysisResults.StdDev, positiveNegativeRatio(result15m1000.ProfitList),
		//	positiveNegativeRatio(result15m500.ProfitList), result15m1000.ProfitList)
	} else {
		strategyAnalysisResults.IsCandidate = false
		//fmt.Printf("no candidata prof1000 %f prof500 %f mean/0.6 %f stddev %f ratio1000 %f ratio500 %f %v\n", result15m1000.Profit, result15m500.Profit,
		//	strategyAnalysisResults.Mean/0.6, strategyAnalysisResults.StdDev, positiveNegativeRatio(result15m1000.ProfitList),
		//	positiveNegativeRatio(result15m500.ProfitList), result15m1000.ProfitList)
	}

	return &strategyAnalysisResults, nil
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

func debugPrint(result analytics.StrategySimulationResult, expected float64) {
	//D if true {
	//D 	fmt.Printf("Best combination last 500 candles: ")
	//D 	for i, _ := range result.Constants {
	//D 		fmt.Printf("Constant%d %.8f ", i+1, result.Constants[i])
	//D 	}
	//D 	fmt.Printf("Profit: %.2f%% Expected %.2f%%\n",
	//D 		result.Profit, expected)
	//D }
}
func stdDev(numbers []float64, mean float64) float64 {
	total := 0.0
	for _, number := range numbers {
		total += math.Pow(number-mean, 2)
	}
	variance := total / float64(len(numbers)-1)
	return math.Sqrt(variance)
}

func sum(numbers []float64) (total float64) {
	for _, x := range numbers {
		total += x
	}
	return total
}

func (mas *MarketAnalysisService) getPairAnalysisResult(pair string) *analytics.PairAnalysis {
	for _, pairStrategyAnalysis := range *mas.PairAnalysisResults {
		if pairStrategyAnalysis.Pair == pair {
			return pairStrategyAnalysis
		}
	}
	return nil
}

func (mas *MarketAnalysisService) getStrategyAnalysisResults(pairAnalysisResult analytics.PairAnalysis, strategy interfaces.Strategy) *analytics.StrategyAnalysis {
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
