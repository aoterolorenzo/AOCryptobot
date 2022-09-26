package analytics

import (
	"gorm.io/gorm"
)

type PairAnalysis struct {
	gorm.Model
	StrategiesAnalysis []StrategyAnalysis `gorm:"foreignKey:PairAnalysisID"`
	TradeSignal        bool
	LockedMonitor      bool
	BestStrategy       string
	MarketDirection    string
	Pair               string
}

type StrategyAnalysis struct {
	gorm.Model
	PairAnalysisID     uint
	StrategyResults    []StrategySimulationResult `gorm:"foreignKey:StrategyAnalysisID"`
	Strategy           string
	IsCandidate        bool
	Mean               float64
	PositivismAvgRatio float64
	StdDev             float64
}

type StrategySimulationResult struct {
	gorm.Model
	StrategyAnalysisID uint
	Period             int
	Profit             float64
	ProfitList         []float64
	Trend              float64
	Constants          []float64
}
