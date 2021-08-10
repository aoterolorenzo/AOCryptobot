package strategies

import (
	"fmt"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"gitlab.com/aoterocom/AOCryptobot/strategies/indicators"
	"reflect"
	"strings"
	"time"
)

type MACDCustomStrategy struct{}

func NewMACDCustomStrategy() MACDCustomStrategy {
	return MACDCustomStrategy{}
}

func (s *MACDCustomStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, []float64{0.10, 0, 0.05})
}

func (s *MACDCustomStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, []float64{0, 1.34})
}

func (s *MACDCustomStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constants []float64) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	lastCandleIndex := len(timeSeries.Candles) - 1

	// Left some margin after the candle start
	if time.Now().Unix()-120 < timeSeries.Candles[lastCandleIndex].Period.Start.Unix() {
		return false
	}

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	currentMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex).Float()
	lastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 1).Float()
	//lastLastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 2).Float()
	myRSI := techan.NewRelativeStrengthIndexIndicator(techan.NewClosePriceIndicator(timeSeries), 12)
	stochRSI := indicators.NewStochasticRelativeStrengthIndicator(myRSI, 12)
	smoothK := techan.NewSimpleMovingAverage(stochRSI, 3)

	lastSmoothKValue := smoothK.Calculate(lastCandleIndex).Float()

	entryRuleSetCheck :=
		currentMACDHistogramValue > lastMACDHistogramValue+constants[0]

	return entryRuleSetCheck && lastSmoothKValue < 50 //&& !(currentMACDHistogramValue < lastMACDHistogramValue - constants[1])
}

func (s *MACDCustomStrategy) ParametrizedShouldExit(timeSeries *techan.TimeSeries, constants []float64) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)
	lastCandleIndex := len(timeSeries.Candles) - 1

	// Left some margin after the candle start
	if time.Now().Unix()-120 < timeSeries.Candles[lastCandleIndex].Period.Start.Unix() {
		return false
	}

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	currentMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex).Float()
	lastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 1).Float()
	myRSI := techan.NewRelativeStrengthIndexIndicator(closePrices, 12)
	currentRSIValue := myRSI.Calculate(lastCandleIndex).Float()

	exitRuleSetCheck := currentMACDHistogramValue < lastMACDHistogramValue-constants[1] || currentRSIValue < 50

	return exitRuleSetCheck && !s.ParametrizedShouldEnter(timeSeries, constants)
}

func (s *MACDCustomStrategy) PerformSimulation(pair string, exchangeService interfaces.ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error) {
	strategyResults := analytics.NewStrategySimulationResult()
	series, err := exchangeService.GetSeries(pair, interval, limit)
	if err != nil {
		return strategyResults, err
	}
	series.Candles = series.Candles[:len(series.Candles)-omit]
	lastCandleIndex := len(series.Candles) - 1
	lastVal := series.Candles[lastCandleIndex].ClosePrice.Float()
	trendPctCondition := lastVal * 0.0006
	jump := lastVal * 0.00002

	highestBalance := -1.0
	highestEnterConstant := -1.0
	highestExitConstant := -1.0

	balance := 1000.0
	var buyRate float64
	var sellRate float64
	open := false
	enterConstant := 0.0
	exitConstant := lastVal * 0.0006
	enterStop := lastVal * 0.0008
	exitStop := lastVal * -0.0002

	selectedEntryConstant := enterConstant
	selectedExitConstant := exitConstant
	var bestProfitList []float64

	if constants != nil {
		selectedExitConstant = (*constants)[1]
		enterConstant = (*constants)[0]
		enterStop = enterConstant + 1
		selectedEntryConstant = enterConstant
		jump = lastVal
	}

	for {
		for ; enterConstant < enterStop; enterConstant += jump {
			balance = 1000.0
			var profitList []float64
			for i := 5; i < len(series.Candles); i++ {

				candles := series.Candles[:i]
				newSeries := series
				newSeries.Candles = candles

				if !open && len(candles) > 4 && s.ParametrizedShouldEnter(&newSeries, []float64{enterConstant, selectedExitConstant, trendPctCondition}) {
					open = true
					buyRate = candles[i-1].ClosePrice.Float()
				} else if open && s.ParametrizedShouldExit(&newSeries, []float64{enterConstant, selectedExitConstant, trendPctCondition}) {
					open = false
					sellRate = candles[i-1].ClosePrice.Float()
					profitPct := sellRate * 1 / buyRate
					balance *= profitPct * (1 - 0.00014)
					profitList = append(profitList, (profitPct*(1-0.0014))-1)
				}
				time.Sleep(100 * time.Microsecond)
			}
			open = false
			//return pairAnalysis, nil

			if balance > highestBalance {
				highestBalance = balance
				selectedEntryConstant = enterConstant
				bestProfitList = profitList
			}
			//fmt.Printf("Entry Constant: %.8f Exit Constant %.8f Balance: %.8f\n", enterConstant, selectedExitConstant, balance)
		}

		if constants != nil {
			break
		}

		var profitList []float64
		for ; exitConstant > exitStop; exitConstant -= jump {
			balance = 1000.0
			for i := 5; i < len(series.Candles); i++ {

				candles := series.Candles[:i]
				newSeries := series
				newSeries.Candles = candles

				if !open && len(candles) > 4 && s.ParametrizedShouldEnter(&newSeries, []float64{selectedEntryConstant, exitConstant, trendPctCondition}) {
					open = true
					buyRate = candles[i-1].ClosePrice.Float()
				} else if open && s.ParametrizedShouldExit(&newSeries, []float64{selectedEntryConstant, exitConstant, trendPctCondition}) {
					open = false
					sellRate = candles[i-1].ClosePrice.Float()
					profitPct := sellRate * 1 / buyRate
					balance *= profitPct * (1 - 0.0014)
					profitList = append(profitList, (profitPct*(1-0.0014))-1)
				}
				time.Sleep(500 * time.Microsecond)
			}
			open = false
			if balance > highestBalance {
				highestBalance = balance
				selectedExitConstant = exitConstant
				bestProfitList = profitList
			}
			//fmt.Printf("Entry Constant: %.8f Exit constant: %.8f Balance: %.8f\n", selectedEntryConstant, exitConstant, balance)
		}

		if selectedEntryConstant != highestEnterConstant && selectedExitConstant != highestExitConstant {
			highestEnterConstant = selectedEntryConstant
			highestExitConstant = selectedExitConstant
		} else {
			//fmt.Printf("BEST CONSTANT COMBINATION FOUND: Entry Constant: %.8f Exit Constant %f Profit: %.4f%%\n",
			//selectedEntryConstant, selectedExitConstant, highestBalance*100/1000-100)
			break
		}
	}

	strategyResults.Profit = highestBalance*100/1000 - 100
	strategyResults.Period = limit - omit
	strategyResults.ProfitList = bestProfitList
	strategyResults.Constants = append(strategyResults.Constants, selectedEntryConstant)
	strategyResults.Constants = append(strategyResults.Constants, selectedExitConstant)
	strategyResults.Constants = append(strategyResults.Constants, trendPctCondition)
	return strategyResults, nil
}

func (s *MACDCustomStrategy) Analyze(pair string, exchangeService interfaces.ExchangeService) (*analytics.StrategyAnalysis, error) {
	strategyAnalysis := analytics.NewStrategyAnalysis()
	strategyAnalysis.Strategy = s

	helpers.Logger.Debugln(fmt.Sprintf("→ Analizing %s",
		strings.Replace(reflect.TypeOf(s).String(), "*strategies.", "", 1)))

	// Analyze last 1000 candles
	result15m1000, err := s.PerformSimulation(pair, exchangeService, "1h", 500, 0, nil)
	if err != nil {
		return nil, err
	}
	// Analyze last 500 candles
	strategyAnalysis.StrategyResults = append(strategyAnalysis.StrategyResults, result15m1000)
	result15m500, err := s.PerformSimulation(pair, exchangeService, "1h", 240, 0, &result15m1000.Constants)
	if err != nil {
		return nil, err
	}
	strategyAnalysis.StrategyResults = append(strategyAnalysis.StrategyResults, result15m500)

	//Calculate profit mean and standard deviation
	profits := []float64{result15m1000.Profit, result15m500.Profit}
	sum := helpers.Sum(profits)
	strategyAnalysis.Mean = sum / float64(len(profits))
	strategyAnalysis.StdDev = helpers.StdDev(profits, strategyAnalysis.Mean)
	strategyAnalysis.PositivismAvgRatio = (helpers.PositiveNegativeRatio(result15m500.ProfitList) + helpers.PositiveNegativeRatio(result15m1000.ProfitList)) / 2

	// Conditions to accept strategy:
	// Conditions to accept strategy:
	if result15m1000.Profit > 3.2 && result15m500.Profit > 2.0 &&
		(helpers.PositiveNegativeRatio(result15m500.ProfitList) >= 1.2 ||
			(len(result15m500.ProfitList) == 0 && helpers.PositiveNegativeRatio(result15m1000.ProfitList) >= 1.2)) {

		strategyAnalysis.IsCandidate = true
		helpers.Logger.Debugln(fmt.Sprintf("✔️  Strategy is tradeable: 1000CandleProfit, %f 500CandleProfit %f, 60%% of the Mean %f, Std Deviation %f, 1000 Profit Ratio %f 500 Profit Ratio %f", result15m1000.Profit, result15m500.Profit,
			strategyAnalysis.Mean/0.6, strategyAnalysis.StdDev, helpers.PositiveNegativeRatio(result15m1000.ProfitList),
			helpers.PositiveNegativeRatio(result15m500.ProfitList)))
	} else {
		strategyAnalysis.IsCandidate = false
		helpers.Logger.Debugln(fmt.Sprintf("❌️ Strategy is NOT tradeable: 1000CandleProfit, %f 500CandleProfit %f, 60%% of the Mean %f, Std Deviation %f, 1000 Profit Ratio %f 500 Profit Ratio %f", result15m1000.Profit, result15m500.Profit,
			strategyAnalysis.Mean/0.6, strategyAnalysis.StdDev, helpers.PositiveNegativeRatio(result15m1000.ProfitList),
			helpers.PositiveNegativeRatio(result15m500.ProfitList)))
	}

	return &strategyAnalysis, nil
}
