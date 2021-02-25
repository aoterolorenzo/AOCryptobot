package services

import "gitlab.com/aoterocom/AOCryptobot/models/analytics"

type PairAnalysisService struct {
	PairStrategyAnalysisList []analytics.PairStrategyAnalysis
}

func (pa *PairAnalysisService) GetPairAnalysisList(pair string) []analytics.PairStrategyAnalysis {
	var pairAnalysisList []analytics.PairStrategyAnalysis
	for _, pairAnalysis := range pa.PairStrategyAnalysisList {
		if pairAnalysis.Pair == pair {
			pairAnalysisList = append(pairAnalysisList, pairAnalysis)
		}
	}
	return pairAnalysisList
}

func (pa *PairAnalysisService) AnalyzePairMarketEntry(pair string) []analytics.PairStrategyAnalysis {
	return nil
}
