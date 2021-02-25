package analytics

type PairStrategyAnalysis struct {
	StrategyResults []StrategyResult
	IsBestStrategy  bool
	Strategy        string
	Pair            string
	TradeSignal     bool
	IsSelected      bool
	Mean            float64
	StdDev          float64
}
