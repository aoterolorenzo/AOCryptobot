package services

import (
	"fmt"
	"github.com/sdcoffey/techan"
	"gitlab.com/aoterocom/AOCryptobot/database"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models/analytics"
	"runtime/debug"
	"time"
)

type MultiMarketService struct {
	SingleMarketServices []*SingleMarketService
	PairAnalysisResults  []*analytics.PairAnalysis
	ExchangeService      *interfaces.ExchangeService
	DatabaseService      *database.DBService
	Interval             string
}

func NewMultiMarketService(databaseService *database.DBService, pairAnalysisResults *[]*analytics.PairAnalysis, interval string) MultiMarketService {
	mms := MultiMarketService{}
	mms.PairAnalysisResults = *pairAnalysisResults
	mms.Interval = interval
	mms.DatabaseService = databaseService

	for _, pairAnalysisResult := range *pairAnalysisResults {
		if !mms.IsMonitoring(pairAnalysisResult.Pair) {
			sms := NewSingleMarketService(databaseService, *techan.NewTimeSeries(), pairAnalysisResult.Pair, interval)
			mms.SingleMarketServices = append(mms.SingleMarketServices, &sms)
		}
	}
	return mms
}

func (mms *MultiMarketService) StartMonitor() {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logger.Errorln(fmt.Sprintf("Recovered. Error on StartMonitor: %+v", r))
			helpers.Logger.Errorln(fmt.Sprintf(string(debug.Stack())))
			time.Sleep(1 * time.Second)
			mms.StartMonitor()
		}
	}()

	for {
		for _, pairAnalysisResult := range mms.PairAnalysisResults {
			isMonitoring := mms.IsMonitoring(pairAnalysisResult.Pair)
			if !isMonitoring {
				mms.startMonitor(pairAnalysisResult.Pair)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (mms *MultiMarketService) ForceMonitor(pair string, databaseService *database.DBService, interval string) {
	sms := NewSingleMarketService(databaseService, *techan.NewTimeSeries(), pair, interval)
	mms.SingleMarketServices = append(mms.SingleMarketServices, &sms)
	go mms.startMonitor(pair)
}

func (mms *MultiMarketService) startMonitor(pair string) {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logger.Errorln(fmt.Sprintf("Recovered. Error on startMonitor: %+v", r))
			helpers.Logger.Errorln(fmt.Sprintf(string(debug.Stack())))
			time.Sleep(1 * time.Second)
			mms.startMonitor(pair)
		}
	}()

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

func (mms *MultiMarketService) SignalAnalyzer() {
	defer func() {
		if r := recover(); r != nil {
			helpers.Logger.Errorln(fmt.Sprintf("Recovered. Error on SignalAnalyzer: %+v", r))
			helpers.Logger.Errorln(fmt.Sprintf(string(debug.Stack())))
			time.Sleep(1 * time.Second)
			mms.SignalAnalyzer()
		}
	}()

	for {
		for _, pairAnalysis := range mms.PairAnalysisResults {
			if pairAnalysis == nil {
				helpers.Logger.Debugln("No Strategies for " + pairAnalysis.Pair)
				continue
			}

			if len(pairAnalysis.StrategiesAnalysis) == 0 {
				//helpers.Logger.Debugln("No PairAnalysisResults for " + pairAnalysis.Pair)
				continue
			}

			for _, strategyAnalysis := range pairAnalysis.StrategiesAnalysis {
				strategy := strategyAnalysis.Strategy.(interfaces.Strategy)
				timeSeries := mms.GetTimeSeries(pairAnalysis.Pair)
				// If no candles continue next
				if len(timeSeries.Candles) < 1 || strategy == nil {
					helpers.Logger.Debugln("No enough candles or no tradeable strategy for " + pairAnalysis.Pair)
					time.Sleep(5 * time.Second)
					continue
				}
				if len(strategyAnalysis.StrategyResults) == 0 {
					helpers.Logger.Debugln("No StrategyResults for " + pairAnalysis.Pair)
					time.Sleep(1 * time.Second)
					continue
				}

				entrySignal := strategy.ParametrizedShouldEnter(timeSeries, strategyAnalysis.StrategyResults[0].Constants)
				exitSignal := strategy.ParametrizedShouldExit(timeSeries, strategyAnalysis.StrategyResults[0].Constants)

				signal := "NEUTRAL"
				if entrySignal == true {
					signal = "ENTRY"
				}
				if exitSignal == true {
					signal = "EXIT"
				}

				mms.DatabaseService.AddSignal(pairAnalysis.Pair, signal, mms.Interval, strategy)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
