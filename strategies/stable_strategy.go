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

type StableStrategy struct{}

func (s *StableStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, []float64{0.15, 0})
}

func (s *StableStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, []float64{0.15, 0})
}

func (s *StableStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constants []float64) bool {

	lastCandleIndex := len(timeSeries.Candles) - 1

	closePrices := techan.NewClosePriceIndicator(timeSeries)
	myRSI := techan.NewRelativeStrengthIndexIndicator(closePrices, 12)

	stochRSI := indicators.NewStochasticRelativeStrengthIndicator(myRSI, 12)
	smoothK := techan.NewSimpleMovingAverage(stochRSI, 3)
	smoothD := techan.NewSimpleMovingAverage(smoothK, 3)

	lastSmoothKValue := smoothK.Calculate(lastCandleIndex).Float()
	lastSmoothDValue := smoothD.Calculate(lastCandleIndex).Float()
	distanceLastKD := lastSmoothKValue - lastSmoothDValue

	lastLastSmoothKValue := smoothK.Calculate(lastCandleIndex - 1).Float()
	lastLastSmoothDValue := smoothD.Calculate(lastCandleIndex - 1).Float()
	distanceLastLastKD := lastLastSmoothKValue - lastLastSmoothDValue

	return distanceLastKD > distanceLastLastKD+constants[0] && myRSI.Calculate(lastCandleIndex).Float() > 40
}

func (s *StableStrategy) ParametrizedShouldExit(timeSeries *techan.TimeSeries, constants []float64) bool {
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

	exitRuleSetCheck := distanceLastKD < distanceLastLastKD || lastRsiValue < lastLastRsiValue

	return exitRuleSetCheck
}

func (s *StableStrategy) PerformSimulation(exchangeService interfaces.ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error) {
	strategyResults := analytics.StrategySimulationResult{}
	series, err := exchangeService.GetSeries(interval, limit)
	if err != nil {
		return strategyResults, err
	}
	series.Candles = series.Candles[:len(series.Candles)-omit]

	balance := 1000.0
	var buyRate float64
	var sellRate float64
	open := false
	entryConstant := 0.1
	entryStop := 0.3
	jump := 0.005
	selectedEntryConstant := 0.0
	highestBalance := -1.0
	var data2 []float64

	for ; entryConstant < entryStop; entryConstant += jump {
		highestBalance = -1.0
		if constants != nil {
			entryConstant = (*constants)[0]
			entryStop = -1.0
		}
		data2 = []float64{}
		balance = 1000.0
		for i := 5; i < len(series.Candles); i++ {

			candles := series.Candles[:i]
			newSeries := series
			newSeries.Candles = candles

			if !open && len(candles) > 4 && s.ParametrizedShouldEnter(&newSeries, []float64{entryConstant, 0}) {
				open = true
				buyRate = candles[i-1].ClosePrice.Float()
				if constants != nil {
					//fmt.Printf("IN %v\n", candles[i-1].Period.End)
				}
			} else if open && s.ParametrizedShouldExit(&newSeries, []float64{entryConstant, 0}) {
				open = false
				sellRate = candles[i-1].ClosePrice.Float()
				profitPct := sellRate * 1 / buyRate
				balance *= profitPct * (1 - 0.0014)
				if constants != nil {
					//fmt.Printf("OUT %v\n", candles[i-1].Period.End)
				}
				data2 = append(data2, profitPct*(1-0.0014)-1)
			}
			time.Sleep(1 * time.Millisecond)
		}

		open = false
		if balance > highestBalance {
			highestBalance = balance
			selectedEntryConstant = entryConstant
		}
		//fmt.Printf("Entry constant: %.8f Balance: %.8f\n", entryConstant, balance)
	}

	//fmt.Println(data2)
	//fmt.Printf("BEST CONSTANT COMBINATION FOUND: Entry Constant: %.8f Profit: %.4f%%\n",
	//	selectedEntryConstant, highestBalance*100/1000-100)

	strategyResults.Trend = series.Candles[len(series.Candles)-1].ClosePrice.Float() / series.Candles[0].ClosePrice.Float()
	strategyResults.Profit = highestBalance*100/1000 - 100
	strategyResults.ProfitList = data2
	strategyResults.Period = limit - omit
	strategyResults.Constants = append(strategyResults.Constants, selectedEntryConstant)
	return strategyResults, nil
}

func (s *StableStrategy) Analyze(exchangeService interfaces.ExchangeService) (*analytics.StrategyAnalysis, error) {
	strategyAnalysis := analytics.StrategyAnalysis{
		IsCandidate: false,
		Strategy:    s,
	}

	helpers.Logger.Debugln(fmt.Sprintf("Analyzing %s\n", strings.Replace(reflect.TypeOf(s).String(), "*strategies.", "", 1)))

	// Analyze last 1000 candles
	result15m1000, err := s.PerformSimulation(exchangeService, "15m", 1000, 0, nil)
	if err != nil {
		return nil, err
	}

	// Analyze last 500 candles
	strategyAnalysis.StrategyResults = append(strategyAnalysis.StrategyResults, result15m1000)
	result15m500, err := s.PerformSimulation(exchangeService, "15m", 500, 0, &result15m1000.Constants)
	if err != nil {
		return nil, err
	}
	strategyAnalysis.StrategyResults = append(strategyAnalysis.StrategyResults, result15m500)

	//Calculate profit mean and standard deviation
	profits := []float64{result15m1000.Profit, result15m500.Profit}
	sum := helpers.Sum(profits)
	strategyAnalysis.Mean = sum / float64(len(profits))
	strategyAnalysis.StdDev = helpers.StdDev(profits, strategyAnalysis.Mean)

	// CONDITIONS: This is a very secure and stable Strategy, so is normal to not have any profit in a 500 or even
	// 100 candles period. Because of that, we will select it as tradeable if it have ALL positive results
	// and even it haven't any result on 500 candles
	if len(result15m1000.ProfitList) > 0 &&
		helpers.AllValuesPositive(result15m1000.ProfitList) &&
		helpers.AllValuesPositive(result15m500.ProfitList) {

		strategyAnalysis.IsCandidate = true
		helpers.Logger.Debugln(fmt.Sprintf("✔️ Strategy is tradeable: 1000CandleProfit, %f 500CandleProfit %f, 60%% of the Mean %f, Std Deviation %f, 1000 Profit Ratio %f 500 Profit Ratio %f\n", result15m1000.Profit, result15m500.Profit,
			strategyAnalysis.Mean/0.6, strategyAnalysis.StdDev, helpers.PositiveNegativeRatio(result15m1000.ProfitList),
			helpers.PositiveNegativeRatio(result15m500.ProfitList)))
	} else {
		strategyAnalysis.IsCandidate = false
		helpers.Logger.Debugln(fmt.Sprintf("❌️ Strategy is NOT tradeable: 1000CandleProfit, %f 500CandleProfit %f, 60%% of the Mean %f, Std Deviation %f, 1000 Profit Ratio %f 500 Profit Ratio %f\n", result15m1000.Profit, result15m500.Profit,
			strategyAnalysis.Mean/0.6, strategyAnalysis.StdDev, helpers.PositiveNegativeRatio(result15m1000.ProfitList),
			helpers.PositiveNegativeRatio(result15m500.ProfitList)))
	}

	return &strategyAnalysis, nil
}
