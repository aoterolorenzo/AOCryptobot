package interfaces

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
)

type Strategy interface {
	ShouldEnter(timeSeries *techan.TimeSeries) bool
	ShouldExit(timeSeries *techan.TimeSeries) bool
	ParametrizedShouldExit(timeSeries *techan.TimeSeries, constant float64) bool
	ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constant float64, trendPct float64) bool
	PerformAnalysis(exchangeService ExchangeService, pair string, constants []float64) (analytics.PairAnalysis, error)
}
