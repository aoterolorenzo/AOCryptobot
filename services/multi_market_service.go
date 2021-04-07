package services

import (
	"fmt"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"reflect"
	"strings"
	"time"
)

type MultiMarketService struct {
	SingleMarketServices []*SingleMarketService
	PairAnalysisResults  []*analytics.PairAnalysis
}

func NewMultiMarketService(pairAnalysisResults *[]*analytics.PairAnalysis) MultiMarketService {
	mms := MultiMarketService{}
	mms.PairAnalysisResults = *pairAnalysisResults

	for _, pairAnalysisResult := range *pairAnalysisResults {
		sms := NewSingleMarketService(*techan.NewTimeSeries(), pairAnalysisResult.Pair)

		mms.SingleMarketServices = append(mms.SingleMarketServices, &sms)
	}

	return mms
}

func (mms *MultiMarketService) StartMonitor() {
	for {
		for _, pairAnalysisResult := range mms.PairAnalysisResults {
			isMonitoring := mms.IsMonitoring(pairAnalysisResult.Pair)
			if pairAnalysisResult.TradeSignal {
				if !isMonitoring {
					helpers.Logger.Infoln(fmt.Sprintf("%s: Monitor started. Strategy %s",
						pairAnalysisResult.Pair, strings.Replace(reflect.TypeOf(pairAnalysisResult.BestStrategy).String(),
							"*strategies.", "", 1)))
					mms.startMonitor(pairAnalysisResult.Pair)
				}
			} else {
				if isMonitoring && !*(*pairAnalysisResult).LockedMonitor {
					helpers.Logger.Infoln(fmt.Sprintf("%s: Monitor stopped", pairAnalysisResult.Pair))
					mms.stopMonitor(pairAnalysisResult.Pair)
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (mms *MultiMarketService) startMonitor(pair string) {
	for _, singleMarketService := range mms.SingleMarketServices {
		if singleMarketService.Pair == pair {
			singleMarketService.StartCandleMonitor(pair)
			return
		}
	}
}

func (mms *MultiMarketService) stopMonitor(pair string) {
	for _, singleMarketService := range mms.SingleMarketServices {
		if singleMarketService.Pair == pair {
			singleMarketService.Active = false
		}
	}
}

func (mms *MultiMarketService) IsMonitoring(pair string) bool {
	for _, singleMarketService := range mms.SingleMarketServices {
		if singleMarketService.Pair == pair {
			return singleMarketService.Active
		}
	}
	return false
}

func (mms *MultiMarketService) GetTimeSeries(pair string) *techan.TimeSeries {
	for _, singleMarketService := range mms.SingleMarketServices {
		if singleMarketService.Pair == pair {
			return &singleMarketService.TimeSeries
		}
	}
	return nil
}
