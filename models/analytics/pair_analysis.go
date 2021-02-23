package analytics

import (
	"github.com/sdcoffey/techan"
)

type PairAnalysis struct {
	Pair            string
	SimulatedProfit float64
	Constants       []float64
	TimeSeries      *techan.TimeSeries
}
