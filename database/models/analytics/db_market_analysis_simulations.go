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
	PairAnalysisID             uint
	StrategySimulationsResults []StrategySimulationResult `gorm:"foreignKey:StrategyAnalysisID"`
	Strategy                   string
	IsCandidate                bool
	Mean                       float64
	PositivismAvgRatio         float64
	StdDev                     float64
}

type StrategySimulationResult struct {
	gorm.Model
	StrategyAnalysisID uint
	Period             int
	Profit             float64
	ProfitList         []StrategySimulationResultProfitList `gorm:"foreignKey:StrategySimulationResultID"`
	Trend              float64
	Constants          []StrategySimulationResultConstant `gorm:"foreignKey:StrategySimulationResultID"`
}

type StrategySimulationResultProfitList struct {
	gorm.Model
	Value                      float64
	StrategySimulationResultID uint
}

type StrategySimulationResultConstant struct {
	gorm.Model
	StrategySimulationResultID uint
	Value                      float64
}
