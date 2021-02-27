package strategies

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"time"
)

type MACDCustomStrategy struct{}

var logger = helpers.Logger{}

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

	entryRuleSetCheck :=
		currentMACDHistogramValue > constants[0] &&
			currentMACDHistogramValue > lastMACDHistogramValue+constants[2]

	return entryRuleSetCheck && !(currentMACDHistogramValue > constants[1] &&
		currentMACDHistogramValue < lastMACDHistogramValue)
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

	exitRuleSetCheck :=
		currentMACDHistogramValue > constants[1] &&
			currentMACDHistogramValue < lastMACDHistogramValue

	exitRuleSetCheck = exitRuleSetCheck && !s.ParametrizedShouldEnter(timeSeries, constants)
	return exitRuleSetCheck
}

func (s *MACDCustomStrategy) PerformAnalysis(exchangeService interfaces.ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error) {
	strategyResults := analytics.StrategySimulationResult{}
	series, err := exchangeService.GetSeries(interval, limit)
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
	enterConstant := lastVal * -0.0002
	exitConstant := lastVal * 0.001
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
				time.Sleep(1 * time.Millisecond)
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
				time.Sleep(1 * time.Millisecond)
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
