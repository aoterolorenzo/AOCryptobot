package strategies

import (
	"fmt"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"gitlab.com/aoterocom/AOCryptobot/strategies/indicators"
	"reflect"
	"strings"
	"time"
)

type Lun1MarCustomStrategy struct {
	Interval string
}

func NewLun1MarCustomStrategy(interval string) Lun1MarCustomStrategy {
	return Lun1MarCustomStrategy{
		Interval: interval,
	}
}

func (s *Lun1MarCustomStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, []float64{0.15, 0})
}

func (s *Lun1MarCustomStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, []float64{0.15, 0})
}

func (s *Lun1MarCustomStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constants []float64) bool {
	lastCandleIndex := len(timeSeries.Candles) - 1

	myRSI := techan.NewRelativeStrengthIndexIndicator(techan.NewClosePriceIndicator(timeSeries), 12)
	stochRSI := indicators.NewStochasticRelativeStrengthIndicator(myRSI, 12)
	smoothK := techan.NewSimpleMovingAverage(stochRSI, 3)
	smoothD := techan.NewSimpleMovingAverage(smoothK, 3)

	//currentRSIValue := myRSI.Calculate(lastCandleIndex).Float()
	//lastRSIValue := myRSI.Calculate(lastCandleIndex-1).Float()
	currentSmoothKValue := smoothK.Calculate(lastCandleIndex).Float()
	currentSmoothDValue := smoothD.Calculate(lastCandleIndex).Float()

	lastSmoothKValue := smoothK.Calculate(lastCandleIndex - 1).Float()
	lastSmoothDValue := smoothD.Calculate(lastCandleIndex - 1).Float()
	distanceCurrentKD := currentSmoothKValue - currentSmoothDValue
	distanceLastKD := lastSmoothKValue - lastSmoothDValue

	myRSI5 := techan.NewRelativeStrengthIndexIndicator(techan.NewClosePriceIndicator(timeSeries), 5)
	myRSI5Value := myRSI5.Calculate(lastCandleIndex).Float()

	//closePrices := techan.NewClosePriceIndicator(timeSeries)
	//MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	//MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)
	//
	//currentMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex).Float()
	//lastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 1).Float()
	//lastLastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 2).Float()

	return currentSmoothKValue > currentSmoothDValue &&
		distanceCurrentKD > distanceLastKD+constants[0] && myRSI5Value < 70

}

func (s *Lun1MarCustomStrategy) ParametrizedShouldExit(timeSeries *techan.TimeSeries, constants []float64) bool {
	myRSI := techan.NewRelativeStrengthIndexIndicator(techan.NewClosePriceIndicator(timeSeries), 12)
	stochRSI := indicators.NewStochasticRelativeStrengthIndicator(myRSI, 12)
	smoothK := techan.NewSimpleMovingAverage(stochRSI, 3)
	smoothD := techan.NewSimpleMovingAverage(smoothK, 3)

	lastCandleIndex := len(timeSeries.Candles) - 1

	lastSmoothKValue := smoothK.Calculate(lastCandleIndex).Float()
	lastSmoothDValue := smoothD.Calculate(lastCandleIndex).Float()
	distanceLastKD := lastSmoothKValue - lastSmoothDValue

	lastLastSmoothKValue := smoothK.Calculate(lastCandleIndex - 1).Float()
	lastLastSmoothDValue := smoothD.Calculate(lastCandleIndex - 1).Float()
	distanceLastLastKD := lastLastSmoothKValue - lastLastSmoothDValue

	lastRsiValue := myRSI.Calculate(lastCandleIndex).Float()
	lastLastRsiValue := myRSI.Calculate(lastCandleIndex - 1).Float()
	exitRuleSetCheck := distanceLastKD < distanceLastLastKD-constants[0] ||
		lastRsiValue < lastLastRsiValue*0.85 ||
		lastSmoothKValue < lastSmoothDValue-0.03

	return exitRuleSetCheck
}

func (s *Lun1MarCustomStrategy) PerformSimulation(pair string, exchangeService interfaces.ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error) {
	strategyResults := analytics.NewStrategySimulationResult()
	series, err := exchangeService.GetSeries(pair, interval, limit)
	if err != nil {
		return strategyResults, err
	}
	series.Candles = series.Candles[:len(series.Candles)-omit]
	highestBalance := -1.0
	balance := 2000.0
	var buyRate float64
	var sellRate float64
	open := false
	entryConstant := 0.0
	entryStop := 0.2
	jump := 0.008
	selectedEntryConstant := 0.0
	var bestProfitList []float64

	if constants != nil {
		entryConstant = (*constants)[0]
	}

	for ; entryConstant < entryStop; entryConstant += jump {

		var profitList []float64
		balance = 2000.0
		for i := 5; i < len(series.Candles); i++ {

			candles := series.Candles[:i]
			newSeries := series
			newSeries.Candles = candles

			if !open && len(candles) > 4 && s.ParametrizedShouldEnter(&newSeries, []float64{entryConstant, 0}) {
				open = true
				buyRate = candles[i-1].ClosePrice.Float()
			} else if open && s.ParametrizedShouldExit(&newSeries, []float64{entryConstant, 0}) {
				open = false
				sellRate = candles[i-1].ClosePrice.Float()
				profitPct := sellRate * 1 / buyRate
				if profitPct < 2 {
					balance *= profitPct * (1 - 0.0014)
					profitList = append(profitList, (profitPct*(1-0.0014))-1)
				}
			}
		}

		open = false
		if balance > highestBalance {
			highestBalance = balance
			selectedEntryConstant = entryConstant
			bestProfitList = profitList
		}

		if constants != nil {
			break
		}
		//fmt.Printf("Entry constant: %.8f Balance: %.8f\n", entryConstant, balance)
	}

	//fmt.Printf("BEST CONSTANT COMBINATION FOUND: Entry Constant: %.8f Profit: %.4f%%\n",
	//	selectedEntryConstant, highestBalance*100/2000-100)

	strategyResults.Trend = series.Candles[len(series.Candles)-1].ClosePrice.Float() / series.Candles[0].ClosePrice.Float()
	strategyResults.Profit = highestBalance*100/2000 - 100
	strategyResults.ProfitList = bestProfitList
	strategyResults.Period = limit - omit
	strategyResults.Constants = append(strategyResults.Constants, selectedEntryConstant)
	return strategyResults, nil
}

func (s *Lun1MarCustomStrategy) Analyze(pair string, exchangeService interfaces.ExchangeService) (*analytics.StrategyAnalysis, error) {
	strategyAnalysis := analytics.NewStrategyAnalysis()
	strategyAnalysis.Strategy = s

	helpers.Logger.Debugln(fmt.Sprintf("→ Analyzing %s", strings.Replace(reflect.TypeOf(s).String(), "*strategies.", "", 1)))

	// Analyze last 2000 candles
	result2000, err := s.PerformSimulation(pair, exchangeService, s.Interval, 2000, 0, nil)
	if err != nil {
		return nil, err
	}
	// Analyze last 500 candles
	strategyAnalysis.StrategyResults = append(strategyAnalysis.StrategyResults, result2000)
	result500, err := s.PerformSimulation(pair, exchangeService, s.Interval, 500, 0, &result2000.Constants)
	if err != nil {
		return nil, err
	}
	strategyAnalysis.StrategyResults = append(strategyAnalysis.StrategyResults, result500)

	sum := helpers.Sum(result2000.ProfitList)
	strategyAnalysis.Mean = sum / float64(len(result2000.ProfitList))
	strategyAnalysis.StdDev = helpers.StdDev(result2000.ProfitList, strategyAnalysis.Mean)
	strategyAnalysis.PositivismAvgRatio = (helpers.PositiveNegativeRatio(result500.ProfitList) + helpers.PositiveNegativeRatio(result2000.ProfitList)) / 2

	timeSeries := techan.TimeSeries{}
	for i, profit := range result2000.ProfitList {
		candle := techan.NewCandle(techan.TimePeriod{Start: time.Unix(int64(i-1), 0), End: time.Unix(int64(i-1), 0)})
		candle.OpenPrice = big.NewDecimal(profit)
		timeSeries.AddCandle(candle)
	}

	lastCandleIndex := len(timeSeries.Candles) - 1
	maIndicator := techan.NewSimpleMovingAverage(techan.NewOpenPriceIndicator(&timeSeries), 5)
	lastMA := maIndicator.Calculate(lastCandleIndex)
	lastlastMA := maIndicator.Calculate(lastCandleIndex - 1)

	// Conditions to accept strategy:
	if result2000.Profit > 3.2 && result500.Profit > 1.0 && lastMA.Float() >= lastlastMA.Float() {

		strategyAnalysis.IsCandidate = true
		helpers.Logger.Debugln(fmt.Sprintf("✅️  Strategy is tradeable: 2000CandleProfit, %f 500CandleProfit %f, Profit Mean %f, Std Deviation %f, 2000 Profit Ratio %f 500 Profit Ratio %f", result2000.Profit, result500.Profit,
			strategyAnalysis.Mean, strategyAnalysis.StdDev, helpers.PositiveNegativeRatio(result2000.ProfitList),
			helpers.PositiveNegativeRatio(result500.ProfitList)))
	} else {
		strategyAnalysis.IsCandidate = false
		helpers.Logger.Debugln(fmt.Sprintf("❌️ Strategy is NOT tradeable: 2000CandleProfit, %f 500CandleProfit %f, Profit Mean %f, Std Deviation %f, 2000 Profit Ratio %f 500 Profit Ratio %f", result2000.Profit, result500.Profit,
			strategyAnalysis.Mean, strategyAnalysis.StdDev, helpers.PositiveNegativeRatio(result2000.ProfitList),
			helpers.PositiveNegativeRatio(result500.ProfitList)))
	}

	return &strategyAnalysis, nil
}
