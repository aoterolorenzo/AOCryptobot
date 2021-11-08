package analytics

import "gitlab.com/aoterocom/AOCryptobot/models"

type PairAnalysis struct {
	StrategiesAnalysis []StrategyAnalysis
	TradeSignal        bool
	LockedMonitor      bool
	BestStrategy       interface{}
	MarketDirection    models.MarketDirection
	Pair               string
}

func NewPairAnalysis(pair string) PairAnalysis {
	return PairAnalysis{
		Pair: pair,
	}
}
