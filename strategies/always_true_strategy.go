package strategies

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
)

type AlwaysTrueStrategy struct{}

func (s *AlwaysTrueStrategy) ShouldEnter(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldEnter(timeSeries, []float64{0.10, 0, 0.05})
}

func (s *AlwaysTrueStrategy) ShouldExit(timeSeries *techan.TimeSeries) bool {
	return s.ParametrizedShouldExit(timeSeries, []float64{0, 1.34})
}

func (s *AlwaysTrueStrategy) ParametrizedShouldEnter(timeSeries *techan.TimeSeries, constants []float64) bool {
	return true
}

func (s *AlwaysTrueStrategy) ParametrizedShouldExit(timeSeries *techan.TimeSeries, constants []float64) bool {
	return true
}

func (s *AlwaysTrueStrategy) PerformSimulation(exchangeService interfaces.ExchangeService, interval string, limit int, omit int, constants *[]float64) (analytics.StrategySimulationResult, error) {
	strategyResults := analytics.NewStrategySimulationResult()
	return strategyResults, nil
}

func (s *AlwaysTrueStrategy) Analyze(exchangeService interfaces.ExchangeService) (*analytics.StrategyAnalysis, error) {
	strategyAnalysis := analytics.NewStrategyAnalysis()
	strategyAnalysis.Strategy = s
	return &strategyAnalysis, nil
}
