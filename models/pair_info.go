package models

type PairInfo struct {
	Max       float64
	Min       float64
	StepSize  float64
	Precision int
}

func NewPairInfo(max float64, min float64, tick float64, precision int) *PairInfo {
	return &PairInfo{
		Max:       max,
		Min:       min,
		StepSize:  tick,
		Precision: precision,
	}
}
