package analytics

type PairAnalysisResult struct {
	strategyAnalysisResults []StrategyAnalysisResult
	TradeSignal             bool
	Pair                    string
}
