package strategies

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
)

type StopLossTriggerStrategy struct{}

func (s *StopLossTriggerStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, []float64{0.10, 0, 0.05})
}

func (s *StopLossTriggerStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, []float64{0, 1.34})
}

func (s *StopLossTriggerStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constants []float64) bool {
	return true
}

func (s *StopLossTriggerStrategy) ParametrizedShouldExit(timeSeries *techan.TimeSeries, constants []float64) bool {
	return true
}

func (s *StopLossTriggerStrategy) PerformSimulation(pair string, exchangeService interfaces.ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error) {
	strategyResults := analytics.NewStrategySimulationResult()
	return strategyResults, nil
}

func (s *StopLossTriggerStrategy) Analyze(pair string, exchangeService interfaces.ExchangeService) (*analytics.StrategyAnalysis, error) {
	strategyAnalysis := analytics.NewStrategyAnalysis()
	strategyAnalysis.Strategy = s
	return &strategyAnalysis, nil
}
