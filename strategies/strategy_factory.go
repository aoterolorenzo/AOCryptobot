package strategies

import (
	"fmt"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"strings"
)

func StrategyFactory(strategyName string, interval string) (interfaces.Strategy, error) {

	switch strings.ToUpper(strategyName) {
	case strings.ToUpper("lun1MarCustomStrategy"):
		lun1MarCustomStrategy := NewLun1MarCustomStrategy(interval)
		return interfaces.Strategy(&lun1MarCustomStrategy), nil
	case strings.ToUpper("lun5JulCustomStrategy"):
		lun5JulCustomStrategy := NewLun5JulCustomStrategy(interval)
		return interfaces.Strategy(&lun5JulCustomStrategy), nil
	case strings.ToUpper("stableStrategy"):
		stableStrategy := NewStableStrategy(interval)
		return interfaces.Strategy(&stableStrategy), nil
	case strings.ToUpper("MACDCustomStrategy"):
		MACDCustomStrategy := NewMACDCustomStrategy(interval)
		return interfaces.Strategy(&MACDCustomStrategy), nil
	case strings.ToUpper("stochRSICustomStrategy"):
		stochRSICustomStrategy := NewStochRSICustomStrategy(interval)
		return interfaces.Strategy(&stochRSICustomStrategy), nil
	case strings.ToUpper("MACDInStochRSIOutCustomStrategy"):
		MACDInStochRSIOutCustomStrategy := NewMACDInStochRSIOutCustomStrategy(interval)
		return interfaces.Strategy(&MACDInStochRSIOutCustomStrategy), nil
	case strings.ToUpper("stochRSIInMACDOutCustomStrategy"):
		stochRSIInMACDOutCustomStrategy := NewStochRSIInMACDOutCustomStrategy(interval)
		return interfaces.Strategy(&stochRSIInMACDOutCustomStrategy), nil
	case strings.ToUpper("mixedStrategy1"):
		mixedStrategy1 := NewMixedStrategy1(interval)
		return interfaces.Strategy(&mixedStrategy1), nil
	case strings.ToUpper("stopLossTriggerStrategy"):
		stopLossTriggerStrategy := StopLossTriggerStrategy{Interval: interval}
		return interfaces.Strategy(&stopLossTriggerStrategy), nil
	default:
		return nil, fmt.Errorf("%s is not a known strategy", strategyName)
	}

}
