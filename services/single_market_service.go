package services

import (
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
	"math"
	"time"
)

type SingleMarketService struct {
	MarketSnapshotsRecord []models.MarketDepth
	TimeSeries            techan.TimeSeries
	Interval              string
	Pair                  string
	Active                bool
}

func NewSingleMarketService(timeSeries techan.TimeSeries, pair string, interval string) SingleMarketService {
	return SingleMarketService{
		MarketSnapshotsRecord: nil,
		Interval:              interval,
		TimeSeries:            timeSeries,
		Pair:                  pair,
		Active:                false,
	}
}

// Get maximum price within the last s seconds
func (sms *SingleMarketService) MaxPrice(s int) (float64, error) {
	if s > len(sms.MarketSnapshotsRecord) {
		s = len(sms.MarketSnapshotsRecord)
	}

	max := 0.0
	for i := 0; i < s && i < len(sms.MarketSnapshotsRecord); i++ {
		if sms.MarketSnapshotsRecord[i].CenterPrice > max {
			max = sms.MarketSnapshotsRecord[i].CenterPrice
		}
	}

	return max, nil
}

// Get minimum price within the last s seconds
func (sms *SingleMarketService) MinPrice(s int) (float64, error) {
	if s > len(sms.MarketSnapshotsRecord) {
		s = len(sms.MarketSnapshotsRecord)
	}

	min := math.MaxFloat64
	for i := 0; i < s && i < len(sms.MarketSnapshotsRecord); i++ {
		if sms.MarketSnapshotsRecord[i].CenterPrice < min {
			min = sms.MarketSnapshotsRecord[i].CenterPrice
		}
	}

	return min, nil
}

func (sms *SingleMarketService) CurrentPricePercentile(s int) (float64, error) {

	maxPrice, err := sms.MaxPrice(s)
	if err != nil {
		return 0.0, err
	}
	minPrice, err := sms.MinPrice(s)
	if err != nil {
		return 0.0, err
	}

	lastPrice := sms.MarketSnapshotsRecord[0].CenterPrice
	spread := maxPrice - minPrice
	aboveMinPrice := lastPrice - minPrice

	return aboveMinPrice * 100 / spread, nil
}

func (sms *SingleMarketService) CurrentPrice() float64 {
	return sms.MarketSnapshotsRecord[0].CenterPrice
}

func (sms *SingleMarketService) PctVariation(s int) (float64, error) {
	if s > len(sms.MarketSnapshotsRecord) {
		s = len(sms.MarketSnapshotsRecord)
	}

	oldPrice := (sms.MarketSnapshotsRecord)[((s - 1) / 2)].CenterPrice
	newPrice := (sms.MarketSnapshotsRecord)[0].CenterPrice

	return 100 - (oldPrice * 100 / newPrice), nil
}

func (sms *SingleMarketService) StartMultiMonitor(pair string) {
	sms.Pair = pair
	sms.Active = true
	binanceService := binance.NewBinanceService()

	go binanceService.DepthMonitor(pair, &sms.MarketSnapshotsRecord)
	go binanceService.TimeSeriesMonitor(pair, sms.Interval, &sms.TimeSeries, &sms.Active)

}

func (sms *SingleMarketService) StartCandleMonitor(pair string) {
	sms.Pair = pair
	sms.Active = true
	binanceService := binance.NewBinanceService()

	go binanceService.TimeSeriesMonitor(pair, sms.Interval, &sms.TimeSeries, &sms.Active)

}

func (sms *SingleMarketService) StopMonitor() {
	sms.Active = false
	time.Sleep(10 * time.Second)
	sms.TimeSeries = *techan.NewTimeSeries()
}
