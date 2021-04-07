package analytics

type StrategyAnalysis struct {
	StrategyResults []StrategySimulationResult
	Strategy        interface{}
	IsCandidate     bool
	Mean            float64
	StdDev          float64
}

func NewStrategyAnalysis() StrategyAnalysis {
	return StrategyAnalysis{
		IsCandidate: false,
	}
}
