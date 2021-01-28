package model

import (
	"github.com/eranyanay/binance-api"
	"strconv"
)

type MarketStatus struct {
	depth          *binance.Depth
	LowerAskPrice  float64
	HigherBidPrice float64
	Spread         float64
	SpreadPct      float64
	CenterPrice    float64
}

func (s *MarketStatus) Set(d *binance.Depth) error {
	s.depth = d
	return s.generateParameters()
}

func (s *MarketStatus) generateParameters() error {
	lowerAskPrice, err := strconv.ParseFloat(s.depth.Asks[0].Price, 64)
	if err != nil {
		return err
	}
	s.LowerAskPrice = lowerAskPrice

	higherBidPrice, err := strconv.ParseFloat(s.depth.Bids[0].Price, 64)
	if err != nil {
		return err
	}
	s.HigherBidPrice = higherBidPrice
	s.Spread = lowerAskPrice - higherBidPrice
	s.CenterPrice = lowerAskPrice - (s.Spread/2)
	s.SpreadPct = s.Spread * 100  / s.CenterPrice

	return nil
}

