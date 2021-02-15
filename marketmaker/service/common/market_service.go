package common

import (
	"github.com/adshao/go-binance/v2"
	"gitlab.com/aoterocom/AOCryptobot/marketmaker/helpers"
	"gitlab.com/aoterocom/AOCryptobot/marketmaker/model"
	"math"
	"reflect"
)

var logger = helpers.Logger{}

type MarketService struct {
	MarketSnapshotsRecord []model.MarketSnapshot
}

// Get maximum price within the last s seconds
func (marketService *MarketService) MaxPrice(s int, marketSnapshot *[]model.MarketSnapshot) (float64, error) {
	if s > len(*marketSnapshot) {
		s = len(*marketSnapshot)
	}

	max := 0.0
	for i := 0; i < s && i < len(*marketSnapshot); i++ {
		if (*marketSnapshot)[i].CenterPrice > max {
			max = (*marketSnapshot)[i].CenterPrice
		}
	}

	return max, nil
}

// Get minimum price within the last s seconds
func (marketService *MarketService) MinPrice(s int, marketSnapshot *[]model.MarketSnapshot) (float64, error) {
	if s > len(*marketSnapshot) {
		s = len(*marketSnapshot)
	}

	min := math.MaxFloat64
	for i := 0; i < s && i < len(*marketSnapshot); i++ {
		if (*marketSnapshot)[i].CenterPrice < min {
			min = (*marketSnapshot)[i].CenterPrice
		}
	}

	return min, nil
}

//TODO: No need marketSnapshot on the arguments since
func (marketService *MarketService) CurrentPricePercentile(s int, marketSnapshot *[]model.MarketSnapshot) (float64, error) {

	maxPrice, err := marketService.MaxPrice(s, marketSnapshot)
	if err != nil {
		return 0.0, err
	}
	minPrice, err := marketService.MinPrice(s, marketSnapshot)
	if err != nil {
		return 0.0, err
	}

	lastPrice := (*marketSnapshot)[0].CenterPrice
	spread := maxPrice - minPrice
	aboveMinPrice := lastPrice - minPrice

	return aboveMinPrice * 100 / spread, nil
}

func (marketService *MarketService) CurrentPrice(marketSnapshot *[]model.MarketSnapshot) float64 {
	return (*marketSnapshot)[0].CenterPrice
}

func (marketService *MarketService) PctVariation(s int, marketSnapshot *[]model.MarketSnapshot) (float64, error) {
	if s > len(*marketSnapshot) {
		s = len(*marketSnapshot)
	}

	oldPrice := (*marketSnapshot)[((s - 1) / 2)].CenterPrice
	newPrice := (*marketSnapshot)[0].CenterPrice

	return 100 - (oldPrice * 100 / newPrice), nil
}

// Adds a model.MarketSnapshot to the record's model.MarketSnapshot
func (marketService *MarketService) AppendStatus(marketSnapshot *model.MarketSnapshot) {
	reverseAny(marketService.MarketSnapshotsRecord)
	marketService.MarketSnapshotsRecord = append(marketService.MarketSnapshotsRecord, *marketSnapshot)
	reverseAny(marketService.MarketSnapshotsRecord)
}

func (marketService *MarketService) StartMonitor(pair string) {
	go marketService.WsDepth(pair)
}

func (marketService *MarketService) WsDepth(pair string) {
	//TODO: Use exchange logic from here and not on market service
	doneC, _, err := binance.WsDepthServe(pair, marketService.WsDepthHandler, marketService.ErrHandler)
	if err != nil {
		logger.Errorln(err)
		return
	}
	<-doneC
}

func (marketService *MarketService) WsDepthHandler(event *binance.WsDepthEvent) {
	marketSnapshot := model.MarketSnapshot{}
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
