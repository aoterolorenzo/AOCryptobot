package strategies

import (
	"github.com/sdcoffey/techan"
)

type MACDStrategy struct {
	timeSeries *techan.TimeSeries
}

func (s *MACDStrategy) SetTimeSeries(timeSeries *techan.TimeSeries) {
	s.timeSeries = timeSeries
}

func (s *MACDStrategy) ShouldEnter() bool {

	closePrices := techan.NewClosePriceIndicator(s.timeSeries)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	constant0 := techan.NewConstantIndicator(0)

	entryRule := techan.NewCrossUpIndicatorRule(constant0, MACDHistogram)

	record := &techan.TradingRecord{}
	return entryRule.IsSatisfied(len(s.timeSeries.Candles)-1, record)
}

func (s *MACDStrategy) ShouldExit() bool {

	closePrices := techan.NewClosePriceIndicator(s.timeSeries)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)
	constant0 := techan.NewConstantIndicator(0)
	exitRule := techan.NewCrossDownIndicatorRule(constant0, MACDHistogram)

	record := &techan.TradingRecord{}
	return exitRule.IsSatisfied(len(s.timeSeries.Candles)-1, record)
}
