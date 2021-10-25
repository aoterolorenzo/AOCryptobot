package database

import "gorm.io/gorm"

// Position is a pair of two Order objects
type Position struct {
	gorm.Model
	Symbol        string  `json:"symbol"`
	EntryTime     int64   `json:"entryTime"`
	ExitTime      int64   `json:"exitTime"`
	Orders        []Order `gorm:"foreignKey:PositionID"`
	Strategy      string
	Constants     []Constant `gorm:"foreignKey:PositionID"`
	Profit        float64
	Gain          float64
	CumulatedGain float64
}

type Constant struct {
	gorm.Model
	PositionID uint
	Value      float64
}
