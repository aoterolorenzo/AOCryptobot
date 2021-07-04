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

type StochRSICustomStrategy struct{}

func NewStochRSICustomStrategy() StochRSICustomStrategy {
	return StochRSICustomStrategy{}
}

func (s *StochRSICustomStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, []float64{0.15, 0})
}

func (s *StochRSICustomStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, []float64{0.15, 0})
}

func (s *StochRSICustomStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constants []float64) bool {
	myRSI := techan.NewRelativeStrengthIndexIndicator(techan.NewClosePriceIndicator(timeSeries), 12)
	stochRSI := indicators.NewStochasticRelativeStrengthIndicator(myRSI, 12)
	smoothK := techan.NewSimpleMovingAverage(stochRSI, 3)
	smoothD := techan.NewSimpleMovingAverage(smoothK, 3)
	lastCandleIndex := len(timeSeries.Candles) - 1

	// Left some margin after the candle start
	if time.Now().Unix()-120 < timeSeries.Candles[lastCandleIndex].Period.Start.Unix() {
		return false
	}

	lastSmoothKValue := smoothK.Calculate(lastCandleIndex).Float()
	lastSmoothDValue := smoothD.Calculate(lastCandleIndex).Float()
	distanceLastKD := lastSmoothKValue - lastSmoothDValue

	lastLastSmoothKValue := smoothK.Calculate(lastCandleIndex - 1).Float()
	lastLastSmoothDValue := smoothD.Calculate(lastCandleIndex - 1).Float()
	distanceLastLastKD := lastLastSmoothKValue - lastLastSmoothDValue

	return ((lastSmoothKValue > lastSmoothDValue+0.1 &&
		distanceLastKD > distanceLastLastKD+constants[0]) ||
		distanceLastKD > 0.16) && lastSmoothKValue < 40
}

func (s *StochRSICustomStrategy) ParametrizedShouldExit(timeSeries *techan.TimeSeries, constants []float64) bool {
	myRSI := techan.NewRelativeStrengthIndexIndicator(techan.NewClosePriceIndicator(timeSeries), 12)
	stochRSI := indicators.NewStochasticRelativeStrengthIndicator(myRSI, 12)
	smoothK := techan.NewSimpleMovingAverage(stochRSI, 3)
	smoothD := techan.NewSimpleMovingAverage(smoothK, 3)

	lastCandleIndex := len(timeSeries.Candles) - 1

	// Left some margin after the candle start
	if time.Now().Unix()-240 < timeSeries.Candles[lastCandleIndex].Period.Start.Unix() {
		return false
	}

	lastSmoothKValue := smoothK.Calculate(lastCandleIndex).Float()
	lastSmoothDValue := smoothD.Calculate(lastCandleIndex).Float()
	distanceLastKD := lastSmoothKValue - lastSmoothDValue

	lastLastSmoothKValue := smoothK.Calculate(lastCandleIndex - 1).Float()
	lastLastSmoothDValue := smoothD.Calculate(lastCandleIndex - 1).Float()
	distanceLastLastKD := lastLastSmoothKValue - lastLastSmoothDValue

	lastRsiValue := myRSI.Calculate(lastCandleIndex).Float()
	lastLastRsiValue := myRSI.Calculate(lastCandleIndex - 1).Float()
	exitRuleSetCheck := distanceLastKD < distanceLastLastKD-0.03 || lastRsiValue < lastLastRsiValue*0.85

	return exitRuleSetCheck && !s.ParametrizedShouldEnter(timeSeries, constants)
}

func (s *StochRSICustomStrategy) PerformSimulation(pair string, exchangeService interfaces.ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error) {
	strategyResults := analytics.NewStrategySimulationResult()
	series, err := exchangeService.GetSeries(pair, interval, limit)
	if err != nil {
		return strategyResults, err
	}
	series.Candles = series.Candles[:len(series.Candles)-omit]

	highestBalance := -1.0
	balance := 1000.0
	var buyRate float64
	var sellRate float64
	open := false
	entryConstant := 0.0
	entryStop := 0.2
	jump := 0.005
	selectedEntryConstant := 0.0
	var bestProfitList []float64

	if constants != nil {
		entryConstant = (*constants)[0]
	}

	for ; entryConstant < entryStop; entryConstant += jump {

		var profitList []float64
		balance = 1000.0
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
				balance *= profitPct * (1 - 0.0014)
				profitList = append(profitList, (profitPct*(1-0.0014))-1)
			}
			time.Sleep(300 * time.Nanosecond)
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
	//	selectedEntryConstant, highestBalance*100/1000-100)

	strategyResults.Trend = series.Candles[len(series.Candles)-1].ClosePrice.Float() / series.Candles[0].ClosePrice.Float()
	strategyResults.Profit = highestBalance*100/1000 - 100
	strategyResults.ProfitList = bestProfitList
	strategyResults.Period = limit - omit
	strategyResults.Constants = append(strategyResults.Constants, selectedEntryConstant)
	return strategyResults, nil
}

func (s *StochRSICustomStrategy) Analyze(pair string, exchangeService interfaces.ExchangeService) (*analytics.StrategyAnalysis, error) {
	strategyAnalysis := analytics.NewStrategyAnalysis()
	strategyAnalysis.Strategy = s

	helpers.Logger.Debugln(fmt.Sprintf("→ Analyzing %s", strings.Replace(reflect.TypeOf(s).String(), "*strategies.", "", 1)))

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

	// Conditions to accept strategy:
	if result15m500.Profit > 2.0 &&
		(helpers.PositiveNegativeRatio(result15m500.ProfitList) >= 1 ||
			(len(result15m500.ProfitList) == 0 && helpers.PositiveNegativeRatio(result15m1000.ProfitList) >= 1.0)) {

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
