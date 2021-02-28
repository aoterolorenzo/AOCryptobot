package interfaces

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
)

type (
	Strategy interface {
		ShouldEnter(timeSeries *techan.TimeSeries) bool
		ShouldExit(timeSeries *techan.TimeSeries) bool
		ParametrizedShouldExit(timeSeries *techan.TimeSeries, constants []float64) bool
		ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constants []float64) bool
		PerformSimulation(exchangeService ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error)
		Analyze(exchangeService ExchangeService) (*analytics.StrategyAnalysis, error)
	}
)
