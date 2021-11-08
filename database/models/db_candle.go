package database

import (
	"github.com/sdcoffey/big"
	"gorm.io/gorm"
)

type Candle struct {
	gorm.Model
	Symbol     string      `json:"symbol" gorm:"uniqueIndex:idx_symbol_period;size:200"`
	Period     string      `json:"period" gorm:"uniqueIndex:idx_symbol_period;size:200"`
	OpenPrice  big.Decimal `json:"openPrice"`
	ClosePrice big.Decimal `json:"closePrice"`
	MaxPrice   big.Decimal `json:"maxPrice"`
	MinPrice   big.Decimal `json:"minPrice"`
	Volume     big.Decimal `json:"volume"`
	TradeCount uint        `json:"tradeCount"`
}
