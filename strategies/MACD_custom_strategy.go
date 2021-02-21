package strategies

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
)

type MACDCustomStrategy struct{}

var logger = helpers.Logger{}

func (s *MACDCustomStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	lastMACDHistogramValue := techan.NewConstantIndicator(MACDHistogram.Calculate(len(timeSeries.Candles) - 2).Float())
	constant0dot15 := techan.NewConstantIndicator(0.15)

	entryRule := techan.And(
		techan.NewCrossUpIndicatorRule(lastMACDHistogramValue, MACDHistogram),
		techan.NewCrossUpIndicatorRule(constant0dot15, MACDHistogram),
	)

	record := &techan.TradingRecord{}

	return entryRule.IsSatisfied(len(timeSeries.Candles)-1, record)
}

func (s *MACDCustomStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	lastMACDHistogramValue := techan.NewConstantIndicator(MACDHistogram.Calculate(len(timeSeries.Candles) - 2).Float())
	constant0dot15 := techan.NewConstantIndicator(0.15)

	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(lastMACDHistogramValue, MACDHistogram),
		techan.NewCrossDownIndicatorRule(constant0dot15, MACDHistogram),
	)

	record := &techan.TradingRecord{}
	return exitRule.IsSatisfied(len(timeSeries.Candles)-1, record)
}
