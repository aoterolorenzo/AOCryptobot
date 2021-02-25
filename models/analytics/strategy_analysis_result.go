package analytics

type StrategyAnalysisResult struct {
	StrategyResults []StrategySimulationResult
	IsBestStrategy  bool
	Strategy        string
	Pair            string
	TradeSignal     bool
	IsSelected      bool
	Mean            float64
	StdDev          float64
}
