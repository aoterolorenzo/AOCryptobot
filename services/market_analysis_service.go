package services

import (
	"fmt"
	"gitlab.com/aoterocom/AOCryptobot/database"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"sort"
	"strings"
	"time"
)

type MarketAnalysisService struct {
	PairAnalysisResults *[]*analytics.PairAnalysis
	ExchangeService     interfaces.ExchangeService
	strategies          []interfaces.Strategy
	dbService           *database.DBService
}

func NewMarketAnalysisService(exchangeService interfaces.ExchangeService,
	strategies []interfaces.Strategy, pairAnalysisResults *[]*analytics.PairAnalysis, dbService *database.DBService) MarketAnalysisService {
	return MarketAnalysisService{
		PairAnalysisResults: pairAnalysisResults,
		ExchangeService:     exchangeService,
		strategies:          strategies,
		dbService:           dbService,
	}
}

func (mas *MarketAnalysisService) PopulateWithPairs(coin string, whitelist []string, blacklist []string) {
	for _, pair := range mas.ExchangeService.GetMarkets(coin, whitelist, blacklist) {
		pairAnalysis := analytics.PairAnalysis{Pair: pair}
		pairAnalysis.TradeSignal = false

		// Set the markets direction
		if strings.HasSuffix(pairAnalysis.Pair, coin) {
			pairAnalysis.MarketDirection = models.MarketDirectionLong
		} else {
			pairAnalysis.MarketDirection = models.MarketDirectionShort
		}

		*mas.PairAnalysisResults = append(*mas.PairAnalysisResults, &pairAnalysis)
	}
}

func (mas *MarketAnalysisService) AnalyzeMarkets() {

	for {
		if *mas.PairAnalysisResults == nil {
			continue
		}

		for _, pairAnalysis := range *mas.PairAnalysisResults {
			pairAnalysisResultPtr := mas.GetPairAnalysisResult(pairAnalysis.Pair)
			newPairAnalysisResult, err := mas.analyzePair(pairAnalysis.Pair)
			newPairAnalysisResult.LockedMonitor = pairAnalysisResultPtr.LockedMonitor
			newPairAnalysisResult.MarketDirection = pairAnalysisResultPtr.MarketDirection
			if err == nil {
				*pairAnalysisResultPtr = newPairAnalysisResult
				if mas.dbService != nil {
					helpers.Logger.Debugln(fmt.Sprintf("Added analisys results: %s ", newPairAnalysisResult.Pair))
					mas.dbService.AddPairAnalysisResult(newPairAnalysisResult)
				}
			}
		}
	}

	//D for _ , item := range mas.GetTradeSignaledAndOpenMarketsByInvStdDev() {
	//D 	fmt.Printf("Best strategy for %s: %s Tradeable: %t.\n", item.Pair, item.BestStrategy,
	//D 		item.TradeSignal)
	//D }
}

func (mas *MarketAnalysisService) analyzePair(pair string) (analytics.PairAnalysis, error) {
	pairAnalysisResult := analytics.PairAnalysis{
		StrategiesAnalysis: nil,
		TradeSignal:        false,
		LockedMonitor:      false,
		Pair:               pair,
	}

	helpers.Logger.Debugln("📐 " + pair + " Market")
	time.Sleep(5 * time.Second)
	// For each strategy in pair
	for _, strategy := range mas.strategies {
		//We analyze the strategy and set the results in pairAnalysisResult
		strategyAnalysisResult, err := strategy.Analyze(pair, mas.ExchangeService)
		if err != nil {
			helpers.Logger.Errorln(err.Error())
			return pairAnalysisResult, err
		}
		pairAnalysisResult.StrategiesAnalysis = append(pairAnalysisResult.StrategiesAnalysis,
			*strategyAnalysisResult)
	}

	//Once the results are set for the pair, we check the strategies and choose between them if they are valid
	chosenStrategy := mas.chooseStrategy(pairAnalysisResult)
	if chosenStrategy != nil && !pairAnalysisResult.LockedMonitor {
		pairAnalysisResult.TradeSignal = true
		pairAnalysisResult.BestStrategy = chosenStrategy
	}

	//

	return pairAnalysisResult, nil
}

func (mas *MarketAnalysisService) chooseStrategy(pairAnalysisResult analytics.PairAnalysis) interface{} {
	var betterStrategy interface{}
	betterStrategy = nil
	bestRatio := -10000.0
	for _, strategy := range pairAnalysisResult.StrategiesAnalysis {
		strategyRatio := (strategy.Mean / strategy.StdDev) * strategy.PositivismAvgRatio
		//D fmt.Printf("Strategy: %s Ratio: %.2f (Analyze %t)\n",strategy.Strategy, strategy.Mean / strategy.StdDev,  strategy.Analyze)
		if strategy.IsCandidate && (strategyRatio > bestRatio || bestRatio == -10000.0) {
			//D fmt.Printf("Strategy %s with ratio %.2f%% is better than %.2f\n", strategy.Strategy,
			//D 	strategy.Mean / strategy.StdDev, bestRatio)
			bestRatio = strategyRatio

			betterStrategy = strategy.Strategy
		} else {
			//D fmt.Printf("Strategy %s with ratio %.2f%% is worst than %.2f\n", strategy.Strategy,
			//D 	strategy.Mean / strategy.StdDev, bestRatio)
		}
	}

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
func (mas *MarketAnalysisService) GetPairAnalysisResult(pair string) *analytics.PairAnalysis {
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

func (mas *MarketAnalysisService) GetTradeSignaledAndOpenMarketsByInvStdDev() []analytics.PairAnalysis {

	pairAnalysisResults := []analytics.PairAnalysis{}
	for _, result := range *mas.PairAnalysisResults {

		if result.TradeSignal || result.LockedMonitor {
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
