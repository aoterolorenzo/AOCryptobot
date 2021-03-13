package bot_signaltrader

import (
	"fmt"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/services"
	"gitlab.com/aoterocom/AOCryptobot/strategies"
	"reflect"
	"strings"
	"time"
)

type Trader struct {
	MarketAnalysisService    *services.MarketAnalysisService
	MultiMarketService       *services.MultiMarketService
	OpenPositions            int
	MaxOpenPositions         int
	enterPrice               map[string]float64
	balance                  float64
	stopLoss                 float64
	tradeQuantityPerPosition float64
	firstExitTriggered       map[string]bool
}

func NewTrader(marketAnalysisService *services.MarketAnalysisService, multiMarketService *services.MultiMarketService) Trader {
	return Trader{
		MarketAnalysisService: marketAnalysisService,
		MultiMarketService:    multiMarketService,
	}
}

func (t *Trader) Start() {

	t.firstExitTriggered = make(map[string]bool)
	t.enterPrice = make(map[string]float64)
	t.balance = 1016.859613
	t.stopLoss = 0.005
	t.MaxOpenPositions = 3
	t.OpenPositions = 0

	for {
		t.tradeQuantityPerPosition = t.balance / 3.04

		for _, pairAnalysisResults := range t.MarketAnalysisService.GetTradeSignaledMarketsByInvStdDev() {
			strategy := pairAnalysisResults.BestStrategy.(interfaces.Strategy)
			pair := pairAnalysisResults.Pair
			timeSeries := t.MultiMarketService.GetTimeSeries(pair)
			results := t.MarketAnalysisService.GetBestStrategyResults(pairAnalysisResults)

			if t.firstExitTriggered[pair] && !pairAnalysisResults.TradeSignal {
				t.firstExitTriggered[pair] = false
			}

			if len(timeSeries.Candles) > 499 && t.enterPrice[pair] > 0 {

				if strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {
					t.LockPair(pair)
					//helpers.Logger.Infoln(fmt.Sprintf("📈 **%s: ❕ Entry signal**\nPrepared doublecheck in 3 minutes", pair))
					go t.DelayedExitCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants, 90)
				} else {
					if timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float() < t.enterPrice[pair]*(1-t.stopLoss) {
						helpers.Logger.Infoln(fmt.Sprintf("📈 **%s: --> Stop Loss **\n", pair))
						t.DelayedExitCheck(pair, &strategies.AlwaysTrueStrategy{}, timeSeries, results.StrategyResults[0].Constants, 0)
					}
				}

			} else if len(timeSeries.Candles) > 499 && t.enterPrice[pair] == 0.0 &&
				t.OpenPositions != t.MaxOpenPositions && t.firstExitTriggered[pair] &&
				strategy.ParametrizedShouldEnter(timeSeries, results.StrategyResults[0].Constants) {

				//helpers.Logger.Infoln(fmt.Sprintf("📈 **%s: ❕ Entry signal**\nPrepared doublecheck in 3 minutes", pair))
				go t.DelayedEntryCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants, 180)
			}

			if !t.firstExitTriggered[pair] && len(timeSeries.Candles) > 499 {
				if strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {
					t.firstExitTriggered[pair] = true
					helpers.Logger.Infoln(
						fmt.Sprintf("%s: Initial exit signal. Time to trade", pair))
				}
			}
		}
		time.Sleep(2 * time.Second)

	}

}

func (t *Trader) LockPair(pair string) {
	for _, marketAnalysisService := range *t.MarketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			*marketAnalysisService.LockedMonitor = true
		}
	}
}

func (t *Trader) UnLockPair(pair string) {
	for _, marketAnalysisService := range *t.MarketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			*marketAnalysisService.LockedMonitor = false
		}
	}
}

func (t *Trader) DelayedEntryCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	if strategy.ParametrizedShouldEnter(timeSeries, constants) && t.enterPrice[pair] == 0.0 &&
		t.OpenPositions != t.MaxOpenPositions && t.firstExitTriggered[pair] {
		t.OpenPositions++
		t.enterPrice[pair] = timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()
		helpers.Logger.Infoln(
			fmt.Sprintf("📈 **%s: ❕ Entry signal**\n", pair) +
				fmt.Sprintf("Strategy: %s\n", strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1)) +
				fmt.Sprintf("Constants: %v\n", constants) +
				fmt.Sprintf("Buy Price: %f\n\n", timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()) +
				fmt.Sprintf("Updated balance: %f", t.balance))
	} else if t.enterPrice[pair] == 0.0 {
		// helpers.Logger.Infoln(fmt.Sprintf("👎🏻 %s: Double entry check fails", pair))
		t.UnLockPair(pair)
	}
}

func (t *Trader) DelayedExitCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	if strategy.ParametrizedShouldExit(timeSeries, constants) && t.enterPrice[pair] > 0 {
		benefit := (t.tradeQuantityPerPosition * timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float() / t.enterPrice[pair]) - t.tradeQuantityPerPosition
		commissionAmount := (t.tradeQuantityPerPosition + benefit) * (0.0007 * 2)
		t.balance += benefit - commissionAmount
		t.tradeQuantityPerPosition += benefit / 3
		t.enterPrice[pair] = 0.0
		profitPct := benefit / 100
		var profitEmoji string
		if profitPct >= 0 {
			profitEmoji = "✅"
		} else {
			profitEmoji = "❌"
		}

		helpers.Logger.Infoln(
			fmt.Sprintf("📉 **%s: ❕ Exit signal**\n", pair) +
				fmt.Sprintf("Strategy: %s\n", strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1)) +
				fmt.Sprintf("Constants: %v\n", constants) +
				fmt.Sprintf("Sell Price: %f\n", timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()) +
				fmt.Sprintf("Updated Balance: %f\n", t.balance) +
				fmt.Sprintf("%s Profit: %f%%", profitEmoji, profitPct))
		t.OpenPositions--
		t.UnLockPair(pair)
	} else {
		//helpers.Logger.Infoln(fmt.Sprintf("👎🏻 %s: Double exit check fails", pair))
	}
}
