package analytics

type PairAnalysis struct {
	StrategiesAnalysis []StrategyAnalysis
	TradeSignal        bool
	LockedMonitor      *bool
	BestStrategy       interface{}
	Pair               string
}
