package bot_signaltrader

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gitlab.com/aoterocom/AOCryptobot/services"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type SignalTraderService struct {
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
	//pairDirection            map[string]models.MarketDirection
}

func NewSignalTrader(marketAnalysisService *services.MarketAnalysisService, multiMarketService *services.MultiMarketService) SignalTraderService {
	return SignalTraderService{
		marketAnalysisService: marketAnalysisService,
		multiMarketService:    multiMarketService,
	}
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/bot_signal-trader/conf.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
}

func (t *SignalTraderService) Start() {

	t.tradingRecordService = services.NewTradingRecordService(t.multiMarketService, t.marketAnalysisService.ExchangeService)
	// Get .env file strategy variables
	t.stopLoss, _ = strconv.ParseFloat(os.Getenv("stopLoss"), 64)
	t.maxOpenPositions, _ = strconv.Atoi(os.Getenv("maxOpenPositions"))
	t.targetCoin = os.Getenv("targetCoin")
	t.tradePctPerPosition, _ = strconv.ParseFloat(os.Getenv("tradePctPerPosition"), 64)
	t.balancePctToTrade, _ = strconv.ParseFloat(os.Getenv("balancePctToTrade"), 64)

	t.firstExitTriggered = make(map[string]bool)
	initialBalance, err := t.marketAnalysisService.ExchangeService.GetAvailableBalance(t.targetCoin)
	if err != nil {
		helpers.Logger.Fatalln(fmt.Sprintf("Couldn't get the initial currentBalance: %s", err.Error()))
	}
	t.initialBalance = initialBalance
	t.currentBalance = initialBalance

	// Infinite loop
	for {
		// For each pair
		for _, pairAnalysisResults := range t.marketAnalysisService.GetTradeSignaledMarketsByInvStdDev() {

			// Update entry amount
			t.tradeQuantityPerPosition = t.currentBalance * t.tradePctPerPosition

			// Set necessary variables from analysis results
			strategy := pairAnalysisResults.BestStrategy.(interfaces.Strategy)
			pair := pairAnalysisResults.Pair
			timeSeries := t.multiMarketService.GetTimeSeries(pair)

			//  If sample is not big enough, continue
			if len(timeSeries.Candles) < 500 {
				continue
			}

			results := t.marketAnalysisService.GetBestStrategyResults(pairAnalysisResults)

			// Makes another first entry mandatory in case analysis becomes inconclusive
			if t.firstExitTriggered[pair] && !pairAnalysisResults.TradeSignal {
				t.firstExitTriggered[pair] = false
			}

			// Scenario 1: Position is open, and we should exit
			if t.ExitCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants) {
				// Exit after a delayed recheck (helps to avoid "false-positive" strategy signals due to market oscillations)
				go t.ExitIfDelayedExitCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants, 180)

				// Scenario 2: Position is closed, and we should enter
			} else if t.EntryCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants) {
				// Entry after a delayed recheck (same as in exit scenario)
				go t.EnterIfDelayedEntryCheck(pair, strategy, timeSeries, results.StrategyResults[0].Constants, 180)
			}

			// Checks just an initial strategy exit signal. This avoid entry in the middle of a market raise
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

func (t *SignalTraderService) EnterIfDelayedEntryCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	if t.EntryCheck(pair, strategy, timeSeries, constants) {
		t.LockPair(pair)
		t.PerformEntry(pair, strategy, timeSeries, constants)
	}
}

func (t *SignalTraderService) ExitIfDelayedExitCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64, delay int) {

	// If there's no stop-loss signal, wait delay and exit if recheck
	time.Sleep(time.Duration(delay) * time.Second)
	if t.ExitCheck(pair, strategy, timeSeries, constants) {
		t.PerformExit(pair, strategy, timeSeries, constants)
		t.UnLockPair(pair)
	}
}

func (t *SignalTraderService) EntryCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64) bool {

	return strategy.ParametrizedShouldEnter(timeSeries, constants) && !t.tradingRecordService.HasOpenPositions(pair) &&
		t.tradingRecordService.OpenPositionsCount() != t.maxOpenPositions && t.firstExitTriggered[pair]
}

func (t *SignalTraderService) ExitCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64) bool {

	return strategy.ParametrizedShouldExit(timeSeries, constants) && t.tradingRecordService.HasOpenPositions(pair)
}

func (t *SignalTraderService) PerformEntry(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64) {

	_ = t.tradingRecordService.EnterPosition(pair, t.tradeQuantityPerPosition,
		t.marketAnalysisService.GetPairAnalysisResult(pair).MarketDirection)
	helpers.Logger.Infoln(
		fmt.Sprintf("📈 **%s: ❕ Entry signal**\n", pair) +
			fmt.Sprintf("Strategy: %s\n", strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1)) +
			fmt.Sprintf("Constants: %v\n", constants) +
			fmt.Sprintf("Buy Price: %f\n\n", timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()) +
			fmt.Sprintf("Updated currentBalance: %f", t.currentBalance))
}

func (t *SignalTraderService) PerformExit(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64) {

	_ = t.tradingRecordService.ExitPositions(pair, t.marketAnalysisService.GetPairAnalysisResult(pair).MarketDirection)

	lastPosition := t.tradingRecordService.LastPosition(pair)

	var profitEmoji string
	profitPct := lastPosition.ProfitPct()

	var tradingAmount float64
	if t.marketAnalysisService.GetPairAnalysisResult(pair).MarketDirection == models.MarketDirectionLong {
		tradingAmount, _ = strconv.ParseFloat(lastPosition.EntranceOrder().CumulativeQuoteQuantity, 64)
	} else {
		tradingAmount, _ = strconv.ParseFloat(lastPosition.EntranceOrder().ExecutedQuantity, 64)
	}

	t.currentBalance += tradingAmount * profitPct

	helpers.Logger.Errorln(fmt.Sprintf("BenefitPct: %f", profitPct))
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
			fmt.Sprintf("Updated Balance: %.2f€\n", t.currentBalance) +
			fmt.Sprintf("Gain/Loss: %.2f€\n", t.currentBalance-t.initialBalance) +
			fmt.Sprintf("%s Profit: %.2f%%", profitEmoji, profitPct*100))
}

func (t *SignalTraderService) LockPair(pair string) {
	for _, marketAnalysisService := range *t.marketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			*marketAnalysisService.LockedMonitor = true
		}
	}
}

func (t *SignalTraderService) UnLockPair(pair string) {
	for _, marketAnalysisService := range *t.marketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			*marketAnalysisService.LockedMonitor = false
		}
	}
}
