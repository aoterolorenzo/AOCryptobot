package services

import (
	"fmt"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"time"
)

type MultiMarketService struct {
	SingleMarketServices []*SingleMarketService
	PairAnalysisResults  []*analytics.PairAnalysis
}

var logger = helpers.Logger{}

func NewMultiMarketService(pairAnalysisResults *[]*analytics.PairAnalysis) MultiMarketService {
	mms := MultiMarketService{}
	mms.PairAnalysisResults = *pairAnalysisResults

	for _, pairAnalysisResult := range *pairAnalysisResults {
		sms := SingleMarketService{
			MarketSnapshotsRecord: nil,
			TimeSeries:            *techan.NewTimeSeries(),
			Pair:                  pairAnalysisResult.Pair,
			Active:                false,
		}

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
					mms.startMonitor(pairAnalysisResult.Pair)
				}
			} else {
				if isMonitoring && !pairAnalysisResult.LockedMonitor {
					mms.stopMonitor(pairAnalysisResult.Pair)
					fmt.Printf("%s: %s Monitor stop\n", time.Now().String(), pairAnalysisResult.Pair)
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (mms *MultiMarketService) startMonitor(pair string) {
	fmt.Printf("%s: Start monitor\n", pair)
	fmt.Printf("%s: %s Monitor start\n", time.Now().String(), pair)
	logger.Infoln(fmt.Sprintf("%s: Start monitor. Estrategia apta detectada\n", pair))
	for _, singleMarketService := range mms.SingleMarketServices {
		if singleMarketService.Pair == pair {
			singleMarketService.StartCandleMonitor(pair)
			return
		}
	}
}

func (mms *MultiMarketService) stopMonitor(pair string) {
	fmt.Printf("%s: Stop monitor\n", pair)
	logger.Infoln(fmt.Sprintf("%s: Stop monitor. La estrategia ya no cumple los requisitos.\n", pair))
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
