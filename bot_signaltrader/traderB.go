package bot_signaltrader

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/services"
	"gitlab.com/aoterocom/AOCryptobot/strategies"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type TraderB struct {
	marketAnalysisService *services.MarketAnalysisService
	multiMarketService    *services.MultiMarketService
	tradingRecordService  *services.TradingRecordService

	maxOpenPositions    int
	targetCoin          string
	stopLoss            float64
	tradePctPerPosition float64
	balancePctToTrade   float64

	currentBalance           float64
	initialBalance           float64
	tradeQuantityPerPosition float64
	firstExitTriggered       map[string]bool
}

func NewTraderB(marketAnalysisService *services.MarketAnalysisService, multiMarketService *services.MultiMarketService) TraderB {
	return TraderB{
		marketAnalysisService: marketAnalysisService,
		multiMarketService:    multiMarketService,
	}
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/bot_signaltrader/conf.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
}

func (t *TraderB) Start() {

	t.tradingRecordService = services.NewTradingRecordService(t.marketAnalysisService.ExchangeService)
	// Get .env file strategy variables
	t.stopLoss, _ = strconv.ParseFloat(os.Getenv("stopLoss"), 64)
	t.maxOpenPositions, _ = strconv.Atoi(os.Getenv("maxOpenPositions"))
	t.targetCoin = os.Getenv("targetCoin")
	t.tradePctPerPosition, _ = strconv.ParseFloat(os.Getenv("tradePctPerPosition"), 64)
	t.balancePctToTrade, _ = strconv.ParseFloat(os.Getenv("balancePctToTrade"), 64)

	t.firstExitTriggered = make(map[string]bool)
	initialBalance, err := t.marketAnalysisService.ExchangeService.GetAvailableBalance(t.targetCoin)
	if err != nil {
		helpers.Logger.Fatalln("Couldn't get the initial currentBalance: %s", err.Error())
	}
	t.initialBalance = initialBalance
	t.currentBalance = initialBalance

	// Infinite loop
	for {

		// Update entry amount
		t.tradeQuantityPerPosition = t.currentBalance * t.tradePctPerPosition

		// For each pair
		for _, pairAnalysisResults := range t.marketAnalysisService.GetTradeSignaledMarketsByInvStdDev() {

			// Set strategy from analysis results
			strategy := pairAnalysisResults.BestStrategy.(interfaces.Strategy)
			pair := pairAnalysisResults.Pair
			timeSeries := t.multiMarketService.GetTimeSeries(pair)
			results := t.marketAnalysisService.GetBestStrategyResults(pairAnalysisResults)

			// Makes another first entry mandatory in case analysis becomes inconclusive
			if t.firstExitTriggered[pair] && !pairAnalysisResults.TradeSignal {
				t.firstExitTriggered[pair] = false
			}

			//  If sample is not big enough, continue
			if len(timeSeries.Candles) < 500 {
				continue
			}

			// If position is open, just check the exit
			if t.tradingRecordService.HasOpenPositions(pair) {

				// If position should Exit, do a delayed re-check (this avoids false positives due the market oscillations)
				if strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {
					t.LockPair(pair)
					//helpers.Logger.Infoln(fmt.Sprintf("📈 **%s: ❕ Entry signal**\nPrepared doublecheck in 3 minutes", pair))
					go t.ExitCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants, 300)

					// Otherwise, manually check that stop loss is not addressed

				} else {
					enterPrice, _ := strconv.
						ParseFloat(t.tradingRecordService.OpenPositions[pair][0].EntranceOrder().Price, 64)
					if timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float() < enterPrice*(1-t.stopLoss) {
						// Stop loss is executed by checking a StopLossTriggerStrategy exit, which results always true
						t.ExitCheck(pair, &strategies.StopLossTriggerStrategy{}, timeSeries, results.StrategyResults[0].Constants, 0)
					}
				}

				// If position is closed, open positions are not the max allowed, first exit was triggered, and position Should enter
			} else if t.tradingRecordService.OpenPositionsCount() != t.maxOpenPositions &&
				t.firstExitTriggered[pair] &&
				strategy.ParametrizedShouldEnter(timeSeries, results.StrategyResults[0].Constants) {

				//helpers.Logger.Infoln(fmt.Sprintf("📈 **%s: ❕ Entry signal**\nPrepared doublecheck in 3 minutes", pair))
				go t.EntryCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants, 180)
			}

			// Checks for a first exit trigger. This avoid entry in the middle of the market raise
			if !t.firstExitTriggered[pair] {
				if strategy.ParametrizedShouldExit(timeSeries, results.StrategyResults[0].Constants) {
					t.firstExitTriggered[pair] = true
					helpers.Logger.Infoln(fmt.Sprintf("%s: Initial exit signal. Time to trade", pair))
				}
			}
		}

		//Wait 2 seconds
		time.Sleep(2 * time.Second)

	}
}

func (t *TraderB) LockPair(pair string) {
	for _, marketAnalysisService := range *t.marketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			*marketAnalysisService.LockedMonitor = true
		}
	}
}

func (t *TraderB) UnLockPair(pair string) {
	for _, marketAnalysisService := range *t.marketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			*marketAnalysisService.LockedMonitor = false
		}
	}
}

func (t *TraderB) EntryCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	lastEnterPrice, _ := strconv.
		ParseFloat(t.tradingRecordService.OpenPositions[pair][0].EntranceOrder().Price, 64)
	if strategy.ParametrizedShouldEnter(timeSeries, constants) && lastEnterPrice == 0.0 &&
		t.tradingRecordService.OpenPositionsCount() != t.maxOpenPositions && t.firstExitTriggered[pair] {

		_ = t.tradingRecordService.EnterPosition(pair, t.tradeQuantityPerPosition)

		helpers.Logger.Infoln(
			fmt.Sprintf("📈 **%s: ❕ Entry signal**\n", pair) +
				fmt.Sprintf("Strategy: %s\n", strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1)) +
				fmt.Sprintf("Constants: %v\n", constants) +
				fmt.Sprintf("Buy Price: %f\n\n", timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()) +
				fmt.Sprintf("Updated currentBalance: %f", t.currentBalance))
	} else if !t.tradingRecordService.HasOpenPositions(pair) {
		// helpers.Logger.Infoln(fmt.Sprintf("👎🏻 %s: Double entry check fails", pair))
		t.UnLockPair(pair)
	}
}

func (t *TraderB) ExitCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	if strategy.ParametrizedShouldExit(timeSeries, constants) && t.tradingRecordService.HasOpenPositions(pair) {

		_ = t.tradingRecordService.ExitPositions(pair)
		lastPosition := t.tradingRecordService.ClosedPositions[pair][len(t.tradingRecordService.ClosedPositions[pair])-1]

		benefit := lastPosition.ProfitPct()
		t.currentBalance += benefit

		var profitEmoji string
		profitPct := lastPosition.ProfitPct()
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
				fmt.Sprintf("Updated Balance: %f\n", t.currentBalance) +
				fmt.Sprintf("%s Profit: %f%%", profitEmoji, profitPct))
		t.UnLockPair(pair)
	} else {
		//helpers.Logger.Infoln(fmt.Sprintf("👎🏻 %s: Double exit check fails", pair))
	}
}
