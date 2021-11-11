package bot_signaltrader

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/database"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gitlab.com/aoterocom/AOCryptobot/services"
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
	databaseService       *database.DBService

	maxOpenPositions int
	targetCoin       string

	// Stop Loss
	stopLoss    bool
	stopLossPct float64

	// Trailing Stop Loss
	trailingStopLoss           bool
	trailingStopLossTriggerPct float64
	trailingStopLossPct        float64
	trailingStopLossArmedAt    map[string]float64

	tradePctPerPosition float64
	balancePctToTrade   float64
	databaseIsEnabled   bool

	currentBalance           float64
	initialBalance           float64
	tradeQuantityPerPosition float64
	firstExitTriggered       map[string]bool
}

func NewSignalTrader(databaseService *database.DBService, marketAnalysisService *services.MarketAnalysisService, multiMarketService *services.MultiMarketService) SignalTraderService {
	return SignalTraderService{
		marketAnalysisService: marketAnalysisService,
		multiMarketService:    multiMarketService,
		databaseService:       databaseService,
	}
}

func NewSignalTraderFullFilled(marketAnalysisService *services.MarketAnalysisService, multiMarketService *services.MultiMarketService, tradingRecordService *services.TradingRecordService, databaseService *database.DBService,
	maxOpenPositions int, targetCoin string, stopLoss bool, stopLossPct float64, trailingStopLoss bool, trailingStopLossTriggerPct float64, trailingStopLossPct float64, trailingStopLossArmedPct map[string]float64,
	tradePctPerPosition float64, balancePctToTrade float64, databaseIsEnabled bool, currentBalance float64, initialBalance float64, tradeQuantityPerPosition float64, firstExitTriggered map[string]bool) SignalTraderService {
	return SignalTraderService{
		marketAnalysisService:      marketAnalysisService,
		multiMarketService:         multiMarketService,
		tradingRecordService:       tradingRecordService,
		databaseService:            databaseService,
		maxOpenPositions:           maxOpenPositions,
		targetCoin:                 targetCoin,
		stopLoss:                   stopLoss,
		stopLossPct:                stopLossPct,
		trailingStopLoss:           trailingStopLoss,
		trailingStopLossTriggerPct: trailingStopLossTriggerPct,
		trailingStopLossPct:        trailingStopLossPct,
		trailingStopLossArmedAt:    trailingStopLossArmedPct,
		tradePctPerPosition:        tradePctPerPosition,
		balancePctToTrade:          balancePctToTrade,
		databaseIsEnabled:          databaseIsEnabled,
		currentBalance:             currentBalance,
		initialBalance:             initialBalance,
		tradeQuantityPerPosition:   tradeQuantityPerPosition,
		firstExitTriggered:         firstExitTriggered,
	}
}

func init() {
	cwd, _ := os.Getwd()
	_ = godotenv.Load(cwd + "/bot_signal-trader/conf.env")
}

func (t *SignalTraderService) Start() {

	t.tradingRecordService = services.NewTradingRecordService(t.multiMarketService, t.marketAnalysisService.ExchangeService)
	// Get .env file strategy variables
	t.stopLoss, _ = strconv.ParseBool(os.Getenv("stopLoss"))
	t.stopLossPct, _ = strconv.ParseFloat(os.Getenv("stopLossPct"), 64)
	t.maxOpenPositions, _ = strconv.Atoi(os.Getenv("maxOpenPositions"))
	t.targetCoin = os.Getenv("targetCoin")
	t.tradePctPerPosition, _ = strconv.ParseFloat(os.Getenv("tradePctPerPosition"), 64)
	t.balancePctToTrade, _ = strconv.ParseFloat(os.Getenv("balancePctToTrade"), 64)
	t.databaseIsEnabled, _ = strconv.ParseBool(os.Getenv("enableDatabaseRecording"))
	t.firstExitTriggered = make(map[string]bool)
	t.trailingStopLoss, _ = strconv.ParseBool(os.Getenv("trailingStopLoss"))
	t.trailingStopLossTriggerPct, _ = strconv.ParseFloat(os.Getenv("trailingStopLossTriggerPct"), 64)
	t.trailingStopLossPct, _ = strconv.ParseFloat(os.Getenv("trailingStopLossPct"), 64)
	t.trailingStopLossArmedAt = make(map[string]float64)
	initialBalance, err := t.marketAnalysisService.ExchangeService.GetAvailableBalance(t.targetCoin)
	if err != nil {
		helpers.Logger.Fatalln(fmt.Sprintf("Couldn't get the initial currentBalance: %s", err.Error()))
	}
	t.initialBalance = initialBalance
	t.currentBalance = initialBalance

	// Infinite loop
	for {
		// For each pair
		for _, pairAnalysisResults := range t.marketAnalysisService.GetTradeSignaledAndOpenMarketsByInvStdDev() {

			// Update entry amount
			t.tradeQuantityPerPosition = t.currentBalance * t.tradePctPerPosition

			// Check results are ready
			if pairAnalysisResults.BestStrategy == nil {
				continue
			}

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
			if t.firstExitTriggered[pair] && !pairAnalysisResults.TradeSignal && !pairAnalysisResults.LockedMonitor {
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

	// If there's no middleChecks exit signal, wait delay and exit if recheck
	time.Sleep(time.Duration(delay) * time.Second)
	if t.ExitCheck(pair, strategy, timeSeries, constants) {
		t.PerformExit(pair, strategy, timeSeries, constants, models.ExitTriggerStrategy)
		t.UnLockPair(pair)
	}
}

func (t *SignalTraderService) MiddleChecks(pair string, timeSeries *techan.TimeSeries) (bool, models.ExitTrigger) {
	entryPrice, _ := strconv.ParseFloat(t.tradingRecordService.OpenPositions[pair][0].EntranceOrder().Price, 64)

	// STOP - LOSS CHECK
	if t.stopLoss {
		if t.StopLossCheck(pair, entryPrice, timeSeries) {
			return true, models.ExitTriggerStopLoss
		}
	}

	// TRIGGER STOP - LOSS CHECK
	if t.trailingStopLoss {
		if t.TrailingStopLossCheck(pair, entryPrice, timeSeries) {
			return true, models.ExitTriggerTrailingStopLoss
		}
	}
	return false, models.ExitTriggerNone
}

func (t *SignalTraderService) StopLossCheck(pair string, entryPrice float64, timeSeries *techan.TimeSeries) bool {
	currentPrice := timeSeries.LastCandle().ClosePrice.Float()

	if t.tradingRecordService.HasOpenPositions(pair) {

		if entryPrice*(1-(t.stopLossPct/100)) > currentPrice {
			helpers.Logger.Debugln(fmt.Sprintf("Stop-Loss signal for %s. Exiting position", pair))
			return true
		}
	}
	return false
}

func (t *SignalTraderService) TrailingStopLossCheck(pair string, entryPrice float64, timeSeries *techan.TimeSeries) bool {
	currentPrice := timeSeries.LastCandle().ClosePrice.Float()

	// Firstly, if price overpass triggerPct, we activate triggerStopLoss
	if currentPrice >= entryPrice*(1+(t.trailingStopLossTriggerPct/100)) && currentPrice > t.trailingStopLossArmedAt[pair] {
		if t.trailingStopLossArmedAt[pair] == 0.0 {
			helpers.Logger.Debugln(fmt.Sprintf("Trailing stop-Loss armed for %s. Current price %f", pair, currentPrice))
		}
		t.trailingStopLossArmedAt[pair] = currentPrice
	}

	// If already triggered
	if t.trailingStopLossArmedAt[pair] != 0.0 {
		targetPrice := t.trailingStopLossArmedAt[pair] * (1 - (t.trailingStopLossPct / 100))
		if targetPrice > currentPrice {
			helpers.Logger.Debugln(fmt.Sprintf("Trailing stop-Loss signal for %s. Target Price %f. Current price %f. Exiting position", pair, targetPrice, currentPrice))
			t.trailingStopLossArmedAt[pair] = 0.0
			t.firstExitTriggered[pair] = false
			return true
		}
	}
	return false
}

func (t *SignalTraderService) EntryCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64) bool {

	return strategy.ParametrizedShouldEnter(timeSeries, constants) && !t.tradingRecordService.HasOpenPositions(pair) &&
		t.tradingRecordService.OpenPositionsCount() != t.maxOpenPositions && t.firstExitTriggered[pair] && !t.IsPairLocked(pair)
}

func (t *SignalTraderService) ExitCheck(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64) bool {

	if t.tradingRecordService.HasOpenPositions(pair) {
		shouldExit, exitTrigger := t.MiddleChecks(pair, timeSeries)
		if shouldExit {
			t.PerformExit(pair, strategy, timeSeries, constants, exitTrigger)
			t.UnLockPair(pair)
			return false
		}

		return strategy.ParametrizedShouldExit(timeSeries, constants)
	}

	return false
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

	lastPosition := t.tradingRecordService.LastOpenPosition(pair)

	if t.databaseIsEnabled {
		lastPosition.Id = t.databaseService.AddPosition(*lastPosition, strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1), constants, -1000, 0.0, models.ExitTriggerNone)
	}
}

func (t *SignalTraderService) PerformExit(pair string, strategy interfaces.Strategy,
	timeSeries *techan.TimeSeries, constants []float64, exitTrigger models.ExitTrigger) {

	_ = t.tradingRecordService.ExitPositions(pair, t.marketAnalysisService.GetPairAnalysisResult(pair).MarketDirection)

	lastPosition := t.tradingRecordService.LastClosedPosition(pair)

	var profitEmoji string
	profitPct := lastPosition.ProfitPct()

	var tradingAmount float64
	if t.marketAnalysisService.GetPairAnalysisResult(pair).MarketDirection == models.MarketDirectionLong {
		tradingAmount, _ = strconv.ParseFloat(lastPosition.EntranceOrder().CumulativeQuoteQuantity, 64)
	} else {
		tradingAmount, _ = strconv.ParseFloat(lastPosition.EntranceOrder().ExecutedQuantity, 64)
	}
	lastCurrentBalance := t.currentBalance
	t.currentBalance += tradingAmount * profitPct / 100

	transactionBenefit := t.currentBalance - lastCurrentBalance

	if profitPct >= 0 {
		profitEmoji = "✅"
	} else {
		profitEmoji = "❌"
	}

	if t.databaseIsEnabled {
		t.databaseService.UpdatePosition(lastPosition.Id, *lastPosition, strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1), constants, profitPct, transactionBenefit, exitTrigger)
	}

	helpers.Logger.Infoln(
		fmt.Sprintf("📉 **%s: ❕ Exit signal**\n", pair) +
			fmt.Sprintf("Strategy: %s\n", strings.Replace(reflect.TypeOf(strategy).String(), "*strategies.", "", 1)) +
			fmt.Sprintf("Trigger: %s\n", exitTrigger) +
			fmt.Sprintf("Constants: %v\n", constants) +
			fmt.Sprintf("Sell Price: %f\n", timeSeries.Candles[len(timeSeries.Candles)-1].ClosePrice.Float()) +
			fmt.Sprintf("Updated Balance: %.2f€\n", t.currentBalance) +
			fmt.Sprintf("Gain/Loss: %.2f€\n", t.currentBalance-t.initialBalance) +
			fmt.Sprintf("%s Profit: %.2f%%", profitEmoji, profitPct))
}

func (t *SignalTraderService) LockPair(pair string) {
	for _, marketAnalysisService := range *t.marketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			marketAnalysisService.LockedMonitor = true
		}
	}
}

func (t *SignalTraderService) UnLockPair(pair string) {
	for _, marketAnalysisService := range *t.marketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			marketAnalysisService.LockedMonitor = false
		}
	}
}

func (t *SignalTraderService) IsPairLocked(pair string) bool {
	for _, marketAnalysisService := range *t.marketAnalysisService.PairAnalysisResults {
		if marketAnalysisService.Pair == pair {
			return marketAnalysisService.LockedMonitor
		}
	}

	return false
}
