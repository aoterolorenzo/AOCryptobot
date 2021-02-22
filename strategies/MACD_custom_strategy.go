package strategies

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"time"
)

type MACDCustomStrategy struct{}

var logger = helpers.Logger{}

func (s *MACDCustomStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, 0.10, 0.05)
}

func (s *MACDCustomStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, 0)
}

func (s *MACDCustomStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constant float64, trendPct float64) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	lastCandleIndex := len(timeSeries.Candles) - 1
	// Check y last candle is about to end
	if time.Now().Unix()+60 < timeSeries.Candles[lastCandleIndex].Period.End.Unix() {
		return false
	}

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
	lastCandleIndex := len(timeSeries.Candles) - 1

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	currentMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex).Float()
	lastMACDHistogramValue := MACDHistogram.Calculate(lastCandleIndex - 1).Float()

	exitRuleSetCheck :=
		currentMACDHistogramValue > constant &&
			currentMACDHistogramValue > lastMACDHistogramValue

	return exitRuleSetCheck
}
