package service

import (
	"../../model"
	"errors"
	"math"
	"reflect"
)

type MarketAnalytics struct {
}

// Get maximum price within the last s seconds
func (analytics *MarketAnalytics) MaxPrice(s int, marketStatusList *[]model.MarketStatus) (float64, error) {
	if s > len(*marketStatusList) {
		return 0.0, errors.New("error: list does still not have enough " +
			"records to cover the requested second period")
	}

	max := 0.0
	for i := 0; i < s && i < len(*marketStatusList); i++ {
		if (*marketStatusList)[i].CenterPrice > max {
			max = (*marketStatusList)[i].CenterPrice
		}
	}

	return max, nil
}

// Get minimum price within the last s seconds
func (analytics *MarketAnalytics) MinPrice(s int, marketStatusList *[]model.MarketStatus) (float64, error) {
	if s > len(*marketStatusList) {
		return 0.0, errors.New("error: list does still not have enough " +
			"records to cover the requested second period")
	}

	min := math.MaxFloat64
	for i := 0; i < s && i < len(*marketStatusList); i++ {
		if (*marketStatusList)[i].CenterPrice < min {
			min = (*marketStatusList)[i].CenterPrice
		}
	}

	return min, nil
}

func (analytics *MarketAnalytics) CurrentPricePercentile(s int, marketStatusList *[]model.MarketStatus) (float64, error)  {

	maxPrice, err := analytics.MaxPrice(s, marketStatusList)
	if err != nil {
		return 0.0, err
	}
	minPrice, err := analytics.MinPrice(s, marketStatusList)
	if err != nil {
		return 0.0, err
	}

	lastPrice := (*marketStatusList)[0].CenterPrice
	spread := maxPrice - minPrice
	aboveMinPrice := lastPrice - minPrice

	return aboveMinPrice * 100 / spread, nil
}

func (analytics *MarketAnalytics) CurrentPrice(marketStatusList *[]model.MarketStatus) float64 {
	return (*marketStatusList)[0].CenterPrice
}

func (analytics *MarketAnalytics) PctVariation(s int, marketStatusList *[]model.MarketStatus) (float64, error)  {
	if s > len(*marketStatusList) {
		return 0.0, errors.New("error: list does still not have enough " +
			"records to cover the requested second period")
	}

	oldPrice := (*marketStatusList)[((s - 1) / 2)].CenterPrice
	newPrice := (*marketStatusList)[0].CenterPrice

	return 100 - (oldPrice * 100 / newPrice), nil
}

func reverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
