package Indicators

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type stochasticRelativeStrengthIndicator struct {
	rsi     techan.Indicator
	minRSI  techan.Indicator
	maxRSI  techan.Indicator
	smoothK int
}

func NewStochasticRelativeStrengthIndicator(baseIndicator techan.Indicator, timeframeStochRSI int) techan.Indicator {
	return stochasticRelativeStrengthIndicator{
		rsi:    baseIndicator,
		minRSI: techan.NewMinimumValueIndicator(baseIndicator, timeframeStochRSI),
		maxRSI: techan.NewMaximumValueIndicator(baseIndicator, timeframeStochRSI),
	}
}

func (srs stochasticRelativeStrengthIndicator) Calculate(index int) big.Decimal {
	rsi := srs.rsi
	minRSI := srs.minRSI
	maxRSI := srs.maxRSI

	dividend := rsi.Calculate(index).Float() - minRSI.Calculate(index).Float()
	divisor := maxRSI.Calculate(index).Float() - minRSI.Calculate(index).Float()

	if divisor == 0.0 {
		return big.NewDecimal(0.0)
	}

	return big.NewDecimal(dividend / divisor)
}
