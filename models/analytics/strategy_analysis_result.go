package analytics

type StrategyAnalysis struct {
	StrategyResults []StrategySimulationResult
	Strategy        interface{}
	IsCandidate     bool
	Mean            float64
	StdDev          float64
}
