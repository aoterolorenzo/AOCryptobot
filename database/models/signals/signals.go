package signals

import "gorm.io/gorm"

type Signal struct {
	gorm.Model
	Pair        string
	Interval    string
	TradeSignal string
	Strategy    string
}
