package strategies

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"time"
)

type MACDCustomStrategy struct{}

var logger = helpers.Logger{}

func (s *MACDCustomStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	// Check y last candle is about to end (last minute), if its not, return false
	if time.Now().Unix()+60 < timeSeries.Candles[len(timeSeries.Candles)-1].Period.End.Unix() {
		return false
	}

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	lastMACDHistogramValue := techan.NewConstantIndicator(MACDHistogram.Calculate(len(timeSeries.Candles) - 2).Float())
	lastlastMACDHistogramValue := techan.NewConstantIndicator(MACDHistogram.Calculate(len(timeSeries.Candles) - 3).Float())
	lastlastlastMACDHistogramValue := techan.NewConstantIndicator(MACDHistogram.Calculate(len(timeSeries.Candles) - 4).Float())
	constant0dot15 := techan.NewConstantIndicator(0.46) // Was 0.1

	entryRule := techan.And(
		techan.NewCrossUpIndicatorRule(lastMACDHistogramValue, MACDHistogram),
		techan.NewCrossUpIndicatorRule(constant0dot15, MACDHistogram),
	)

	entryRule = techan.And(
		entryRule,
		techan.NewCrossUpIndicatorRule(lastlastMACDHistogramValue, MACDHistogram),
	)

	entryRule = techan.And(
		entryRule,
		techan.NewCrossUpIndicatorRule(lastlastlastMACDHistogramValue, MACDHistogram),
	)

	record := &techan.TradingRecord{}

	entryCondition := MACDHistogram.Calculate(len(timeSeries.Candles)-1).Float() >
		MACDHistogram.Calculate(len(timeSeries.Candles)-2).Float()+0.05

	return entryRule.IsSatisfied(len(timeSeries.Candles)-1, record) && entryCondition

}

func (s *MACDCustomStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(timeSeries)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	MACDHistogram := techan.NewMACDHistogramIndicator(MACD, 9)

	//lastValue := MACDHistogram.Calculate(len(timeSeries.Candles) - 2).Float()
	lastMACDHistogramValue := techan.NewConstantIndicator(MACDHistogram.Calculate(len(timeSeries.Candles) - 2).Float())
	constant0dot15 := techan.NewConstantIndicator(1.20)

	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(lastMACDHistogramValue, MACDHistogram),
		techan.NewCrossDownIndicatorRule(constant0dot15, MACDHistogram),
	)

	record := &techan.TradingRecord{}
	return exitRule.IsSatisfied(len(timeSeries.Candles)-1, record)
}
