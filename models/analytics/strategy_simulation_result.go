package analytics

type StrategySimulationResult struct {
	Period     int
	Profit     float64
	ProfitList []float64
	Trend      float64
	Constants  []float64
}

func NewStrategySimulationResult() StrategySimulationResult {
	return StrategySimulationResult{}
}
