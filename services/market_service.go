package services

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
	"math"
)

type MarketService struct {
	MarketSnapshotsRecord []models.MarketDepth
	TimeSeries            techan.TimeSeries
}

// Get maximum price within the last s seconds
func (marketService *MarketService) MaxPrice(s int) (float64, error) {
	if s > len(marketService.MarketSnapshotsRecord) {
		s = len(marketService.MarketSnapshotsRecord)
	}

	max := 0.0
	for i := 0; i < s && i < len(marketService.MarketSnapshotsRecord); i++ {
		if marketService.MarketSnapshotsRecord[i].CenterPrice > max {
			max = marketService.MarketSnapshotsRecord[i].CenterPrice
		}
	}

	return max, nil
}

// Get minimum price within the last s seconds
func (marketService *MarketService) MinPrice(s int) (float64, error) {
	if s > len(marketService.MarketSnapshotsRecord) {
		s = len(marketService.MarketSnapshotsRecord)
	}

	min := math.MaxFloat64
	for i := 0; i < s && i < len(marketService.MarketSnapshotsRecord); i++ {
		if marketService.MarketSnapshotsRecord[i].CenterPrice < min {
			min = marketService.MarketSnapshotsRecord[i].CenterPrice
		}
	}

	return min, nil
}

func (marketService *MarketService) CurrentPricePercentile(s int) (float64, error) {

	maxPrice, err := marketService.MaxPrice(s)
	if err != nil {
		return 0.0, err
	}
	minPrice, err := marketService.MinPrice(s)
	if err != nil {
		return 0.0, err
	}

	lastPrice := marketService.MarketSnapshotsRecord[0].CenterPrice
	spread := maxPrice - minPrice
	aboveMinPrice := lastPrice - minPrice

	return aboveMinPrice * 100 / spread, nil
}

func (marketService *MarketService) CurrentPrice() float64 {
	return marketService.MarketSnapshotsRecord[0].CenterPrice
}

func (marketService *MarketService) PctVariation(s int) (float64, error) {
	if s > len(marketService.MarketSnapshotsRecord) {
		s = len(marketService.MarketSnapshotsRecord)
	}

	oldPrice := (marketService.MarketSnapshotsRecord)[((s - 1) / 2)].CenterPrice
	newPrice := (marketService.MarketSnapshotsRecord)[0].CenterPrice

	return 100 - (oldPrice * 100 / newPrice), nil
}

func (marketService *MarketService) StartMonitor(pair string) {
	binanceService := binance.BinanceService{}
	binanceService.ConfigureClient()
	binanceService.SetPair(pair)

	go binanceService.DepthMonitor(&marketService.MarketSnapshotsRecord)
	go binanceService.TimeSeriesMonitor("15m", &marketService.TimeSeries)

}
