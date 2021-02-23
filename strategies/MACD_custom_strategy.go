package strategies

import (
	"fmt"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
)

type MACDCustomStrategy struct{}

var logger = helpers.Logger{}

func (s *MACDCustomStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, 0.10, 0.05)
}

func (s *MACDCustomStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, 1.34)
}

func (s *MACDCustomStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constant float64, trendPct float64) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	lastCandleIndex := len(timeSeries.Candles) - 1
	// Check y last candle is about to end
	/*if time.Now().Unix()+60 < timeSeries.Candles[lastCandleIndex].Period.End.Unix() {
		return false
	}*/

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	currentMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex).Float()
	lastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 1).Float()
	lastLastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 2).Float()

	entryRuleSetCheck :=
		currentMACDHistogramValue > constant &&
			currentMACDHistogramValue > lastMACDHistogramValue+trendPct &&
			lastMACDHistogramValue > lastLastMACDHistogramValue+trendPct

	return entryRuleSetCheck
}

func (s *MACDCustomStrategy) ParametrizedShouldExit(timeSeries *techan.TimeSeries, constant float64) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	lastMACDHistogramValue := techan.NewConstantIndicator(MACDHistogram.Calculate(len(timeSeries.Candles) - 2).Float())
	constant0dot15 := techan.NewConstantIndicator(constant)

	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(lastMACDHistogramValue, MACDHistogram),
		techan.NewCrossDownIndicatorRule(constant0dot15, MACDHistogram),
	)

	record := &techan.TradingRecord{}
	return exitRule.IsSatisfied(len(timeSeries.Candles)-1, record)
}

func (s *MACDCustomStrategy) PerformAnalysis(exchangeService interfaces.ExchangeService, pair string, interval string, constants []float64) (analytics.PairAnalysis, error) {
	pairAnalysis := analytics.PairAnalysis{}
	series, err := exchangeService.GetSeries(interval)
	if err != nil {
		return pairAnalysis, err
	}

	lastCandleIndex := len(series.Candles) - 1
	lastVal := series.Candles[lastCandleIndex].ClosePrice.Float()
	trendPctCondition := lastVal * 0.000034
	jump := lastVal * 0.00002

	highestBalance := -1.0
	highestEnterConstant := -1.0
	highestExitConstant := -1.0
	var selectedEntryConstant float64
	var selectedExitConstant float64

	balance := 1000.0
	var buyRate float64
	var sellRate float64
	open := false

	for {
		enterConstant := lastVal * -0.0002
		exitConstant := lastVal * 0.001
		enterStop := lastVal * 0.001
		exitStop := lastVal * -0.0002

		for ; enterConstant < enterStop; enterConstant += jump {
			balance = 1000.0
			for i := 4; i < len(series.Candles); i++ {
				candles := series.Candles[:i]
				newSeries := series
				newSeries.Candles = candles

				if !open && len(candles) > 4 && s.ParametrizedShouldEnter(&newSeries, enterConstant, trendPctCondition) {
					open = true
					buyRate = candles[i-1].ClosePrice.Float()
				} else if open && s.ParametrizedShouldExit(&newSeries, selectedExitConstant) {
					open = false
					sellRate = candles[i-1].ClosePrice.Float()
					profitPct := sellRate * 1 / buyRate
					balance *= profitPct * (1 - 0.00014)
				}
			}

			if balance > highestBalance {
				highestBalance = balance
				selectedEntryConstant = enterConstant
			}
			fmt.Printf("Entry Constant: %.8f Exit Constant %.8f Balance: %.8f\n", enterConstant, exitConstant, balance)
		}

		for ; exitConstant > exitStop; exitConstant -= jump {
			balance = 1000.0
			for i := 4; i < len(series.Candles); i++ {

				candles := series.Candles[:i]
				newSeries := series
				newSeries.Candles = candles

				if !open && len(candles) > 4 && s.ParametrizedShouldEnter(&newSeries, selectedEntryConstant, trendPctCondition) {
					open = true
					buyRate = candles[i-1].ClosePrice.Float()
				} else if open && s.ParametrizedShouldExit(&newSeries, exitConstant) {
					open = false
					sellRate = candles[i-1].ClosePrice.Float()
					profitPct := sellRate * 1 / buyRate
					balance *= profitPct * (1 - 0.00014)
				}
			}
			if balance > highestBalance {
				highestBalance = balance
				selectedExitConstant = exitConstant
			}
			fmt.Printf("Entry Constant: %.8f Exit constant: %.8f Balance: %.8f\n", selectedEntryConstant, exitConstant, balance)
		}

		if selectedEntryConstant != highestEnterConstant && selectedExitConstant != highestExitConstant {
			highestEnterConstant = selectedEntryConstant
			highestExitConstant = selectedExitConstant
		} else {
			fmt.Printf("%s: BEST CONSTANT COMBINATION FOUND: Entry Constant: %.8f Exit Constant %.8f Profit: %.4f%%\n",
				pair, selectedEntryConstant, selectedExitConstant, highestBalance*100/1000-100)
			break
		}
	}

	pairAnalysis.SimulatedProfit = highestBalance*100/1000 - 100
	pairAnalysis.Constants = append(pairAnalysis.Constants, selectedEntryConstant)
	pairAnalysis.Constants = append(pairAnalysis.Constants, selectedExitConstant)

	return pairAnalysis, nil
}
