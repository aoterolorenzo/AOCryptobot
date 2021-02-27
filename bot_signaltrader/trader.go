package bot_signaltrader

import (
	"fmt"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
	"gitlab.com/aoterocom/AOCryptobot/services"
	"reflect"
	"strings"
	"time"
)

type Trader struct {
	MarketAnalysisService *services.MarketAnalysisService
	MultiMarketService    *services.MultiMarketService
}

func NewTrader(marketAnalysisService *services.MarketAnalysisService, multiMarketService *services.MultiMarketService) Trader {
	return Trader{
		MarketAnalysisService: marketAnalysisService,
		MultiMarketService:    multiMarketService,
	}
}

func (t *Trader) Start() {

	firstExitTriggered := make(map[string]bool)
	enterPrice := make(map[string]float64)
	balance := 1006.26931
	tradeQuantityPerPosition := 330.0
	maxOpenPositions := 3
	openPositions := 0

	for {
		for _, pairAnalysisResults := range t.MarketAnalysisService.GetTradeSignaledMarketsByInvStdDev() {
			strategy := pairAnalysisResults.BestStrategy.(interfaces.Strategy)
			pair := pairAnalysisResults.Pair
			timeSeries := t.MultiMarketService.GetTimeSeries(pair)
			results := t.MarketAnalysisService.GetBestStrategyResults(pairAnalysisResults)
			timeNow := time.Now().String()

			if len(timeSeries.Candles) > 499 &&
				(time.Now().Unix() < timeSeries.Candles[len(timeSeries.Candles)-1].Period.End.Unix()-180 ||
					time.Now().Unix() < timeSeries.Candles[len(timeSeries.Candles)-1].Period.Start.Unix()+180) {
				continue
			}

			if enterPrice[pair] == 0.0 &&
				openPositions != maxOpenPositions &&
				len(timeSeries.Candles) > 499 &&
				firstExitTriggered[pair] &&
				strategy.ParametrizedShouldEnter(timeSeries, results.StrategyResults[0].Constants) {

				t.LockPair(pair)
				openPositions++
				bs := binance.BinanceService{}
				bs.ConfigureClient()
				bs.SetPair(pairAnalysisResults.Pair)
				// Left some margin after the candle start
				// Left some margin after the candle start
				enterPrice[pair] = timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()
				fmt.Printf("%s: %s ! Entry signal\n", timeNow, pair)
				fmt.Printf("%s: %s Strategy %s\n", timeNow, pair, strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1))
				fmt.Printf("%s: %s Constants %v\n", timeNow, pair, results.StrategyResults[0].Constants)
				fmt.Printf("%s: %s Price %f\n\n", timeNow, pair, timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float())

			} else if enterPrice[pair] > 0 && len(timeSeries.Candles) > 499 && strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {

				benefit := (tradeQuantityPerPosition * timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float() / enterPrice[pair]) - tradeQuantityPerPosition
				balance += benefit
				balance *= 1 - 0.0014
				tradeQuantityPerPosition += benefit / 3
				enterPrice[pair] = 0.0
				fmt.Printf("%s: %s ! Exit signal\n", timeNow, pair)
				fmt.Printf("%s: %s Strategy %s\n", timeNow, pair, strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1))
				fmt.Printf("%s: %s Constants %v\n", timeNow, pair, results.StrategyResults[0].Constants)
				fmt.Printf("%s: %s Price %f\n", timeNow, pair, timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float())
				fmt.Printf("%s: %s Updated balance %f\n", timeNow, pair, balance)
				fmt.Printf("%s: %s Profit %f%%\n\n", timeNow, pair, benefit/100)
				t.UnLockPair(pair)
				openPositions--
			}

			if !firstExitTriggered[pair] && len(timeSeries.Candles) > 499 {
				if strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {
					firstExitTriggered[pair] = true
					fmt.Printf("%s: %s First exit triggered\n", timeNow, pair)
				}
			}

		}
		time.Sleep(2 * time.Second)

	}

}

func (t *Trader) LockPair(pair string) {
	for _, marketAnalysisService := range *t.MarketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			marketAnalysisService.LockedMonitor = true
		}
	}
}

func (t *Trader) UnLockPair(pair string) {
	for _, marketAnalysisService := range *t.MarketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			marketAnalysisService.LockedMonitor = true
		}
	}
}
