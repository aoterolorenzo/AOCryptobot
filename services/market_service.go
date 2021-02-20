package services

import (
	"github.com/adshao/go-binance/v2"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"math"
	"reflect"
)

var logger = helpers.Logger{}

type MarketService struct {
	MarketSnapshotsRecord []models.MarketSnapshot
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

// Adds a models.MarketSnapshot to the record's models.MarketSnapshot
func (marketService *MarketService) AppendStatus(marketSnapshot *models.MarketSnapshot) {
	reverseAny(marketService.MarketSnapshotsRecord)
	marketService.MarketSnapshotsRecord = append(marketService.MarketSnapshotsRecord, *marketSnapshot)
	reverseAny(marketService.MarketSnapshotsRecord)
}

func (marketService *MarketService) StartMonitor(pair string) {
	go marketService.WsDepth(pair)
}

func (marketService *MarketService) WsDepth(pair string) {
	//TODO: Use exchange logic from here and not on market services
	doneC, _, err := binance.WsDepthServe(pair, marketService.WsDepthHandler, marketService.ErrHandler)
	if err != nil {
		logger.Errorln(err)
		return
	}
	<-doneC
}

func (marketService *MarketService) WsDepthHandler(event *binance.WsDepthEvent) {
	marketSnapshot := models.MarketSnapshot{}
	err := marketSnapshot.Set(event)
	if err != nil {
		marketService.ErrHandler(err)
		return
	}

	// Grab the record only if there are 2 bid and ask groups
	if marketSnapshot.WsDepthEvent != nil && marketSnapshot.WsDepthEvent.Bids != nil && marketSnapshot.WsDepthEvent.Asks != nil &&
		len(marketSnapshot.WsDepthEvent.Bids) > 2 && len(marketSnapshot.WsDepthEvent.Asks) > 2 {
		marketService.AppendStatus(&marketSnapshot)
	}
}

func (marketService *MarketService) ErrHandler(err error) {
	logger.Errorln(err)
}

func reverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
