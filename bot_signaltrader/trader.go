package bot_signaltrader

import (
	"fmt"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
	"gitlab.com/aoterocom/AOCryptobot/services"
	"reflect"
	"strings"
	"time"
)

type Trader struct {
	MarketAnalysisService *services.MarketAnalysisService
	MultiMarketService    *services.MultiMarketService
	OpenPositions         *[]models.Position
	MaxOpenPositions      int
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
	candleCheck := make(map[string]techan.TimePeriod)
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

			if firstExitTriggered[pair] && !pairAnalysisResults.TradeSignal {
				firstExitTriggered[pair] = false
			}

			if len(timeSeries.Candles) > 499 &&
				enterPrice[pair] > 0 &&
				candleCheck[pair] != timeSeries.Candles[len(timeSeries.Candles)-1].Period &&
				strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {

				benefit := (tradeQuantityPerPosition * timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float() / enterPrice[pair]) - tradeQuantityPerPosition
				balance += benefit
				balance *= 1 - 0.0014
				tradeQuantityPerPosition += benefit / 3
				enterPrice[pair] = 0.0
				profitPct := benefit/100 - 1
				var profitEmoji string
				if profitPct >= 0 {
					profitEmoji = "✅"
				} else {
					profitEmoji = "❌"
				}

				helpers.Logger.Infoln(
					fmt.Sprintf("📉 **%s: ❕ Exit signal**\n", pair) +
						fmt.Sprintf("Strategy: %s\n", strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1)) +
						fmt.Sprintf("Constants: %v\n", results.StrategyResults[0].Constants) +
						fmt.Sprintf("Sell Price: %f\n", timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()) +
						fmt.Sprintf("Updated Balance: %f\n", balance) +
						fmt.Sprintf("%s Profit: %f%%", profitEmoji, profitPct))
				t.UnLockPair(pair)
				candleCheck[pair] = timeSeries.Candles[len(timeSeries.Candles)-1].Period

				openPositions--
			} else if len(timeSeries.Candles) > 499 &&
				enterPrice[pair] == 0.0 &&
				candleCheck[pair] != timeSeries.Candles[len(timeSeries.Candles)-1].Period &&
				openPositions != maxOpenPositions &&
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
				helpers.Logger.Infoln(
					fmt.Sprintf("📈 **%s: ❕ Entry signal**\n", pair) +
						fmt.Sprintf("Strategy: %s\n", strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1)) +
						fmt.Sprintf("Constants: %v\n", results.StrategyResults[0].Constants) +
						fmt.Sprintf("Buy Price: %f\n\n", timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()) +

						fmt.Sprintf("Updated balance: %f", balance))
				candleCheck[pair] = timeSeries.Candles[len(timeSeries.Candles)-1].Period

			}

			if !firstExitTriggered[pair] && len(timeSeries.Candles) > 499 {
				if strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {
					firstExitTriggered[pair] = true
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

func (t *Trader) InitPosition(pair string) {}
