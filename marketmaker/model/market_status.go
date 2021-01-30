package model

import (
	"github.com/adshao/go-binance/v2"
	"strconv"
)

type MarketStatus struct {
	WsDepthEvent   *binance.WsDepthEvent
	LowerAskPrice  float64
	HigherBidPrice float64
	Spread         float64
	SpreadPct      float64
	CenterPrice    float64
}

func (s *MarketStatus) Set(depthEvent *binance.WsDepthEvent) error {
	if len(depthEvent.Asks) > 0 && len(depthEvent.Bids) > 0 {
		s.WsDepthEvent = depthEvent
		return s.generateParameters()
	}
	return nil
}

func (s *MarketStatus) generateParameters() error {
	lowerAskPrice, err := strconv.ParseFloat(s.WsDepthEvent.Asks[0].Price, 64)
	if err != nil {
		return err
	}
	s.LowerAskPrice = lowerAskPrice

	higherBidPrice, err := strconv.ParseFloat(s.WsDepthEvent.Bids[0].Price, 64)
	if err != nil {
		return err
	}
	s.HigherBidPrice = higherBidPrice

	s.Spread = lowerAskPrice - higherBidPrice
	s.CenterPrice = lowerAskPrice - (s.Spread / 2)
	s.SpreadPct = s.Spread * 100 / s.CenterPrice

	return nil
}
