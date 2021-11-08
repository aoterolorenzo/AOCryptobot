package strategies

import (
	"fmt"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
)

func StrategyFactory(strategyName string, interval string) (interfaces.Strategy, error) {

	switch strategyName {
	case "lun1MarCustomStrategy":
		lun1MarCustomStrategy := NewLun1MarCustomStrategy(interval)
		return interfaces.Strategy(&lun1MarCustomStrategy), nil
	case "lun5JulCustomStrategy":
		lun5JulCustomStrategy := NewLun5JulCustomStrategy(interval)
		return interfaces.Strategy(&lun5JulCustomStrategy), nil
	case "stableStrategy":
		stableStrategy := NewStableStrategy(interval)
		return interfaces.Strategy(&stableStrategy), nil
	case "MACDCustomStrategy":
		MACDCustomStrategy := NewMACDCustomStrategy(interval)
		return interfaces.Strategy(&MACDCustomStrategy), nil
	case "stochRSICustomStrategy":
		stochRSICustomStrategy := NewStochRSICustomStrategy(interval)
		return interfaces.Strategy(&stochRSICustomStrategy), nil
	case "MACDInStochRSIOutCustomStrategy":
		MACDInStochRSIOutCustomStrategy := NewMACDInStochRSIOutCustomStrategy(interval)
		return interfaces.Strategy(&MACDInStochRSIOutCustomStrategy), nil
	case "stochRSIInMACDOutCustomStrategy":
		stochRSIInMACDOutCustomStrategy := NewStochRSIInMACDOutCustomStrategy(interval)
		return interfaces.Strategy(&stochRSIInMACDOutCustomStrategy), nil
	case "mixedStrategy1":
		mixedStrategy1 := NewMixedStrategy1(interval)
		return interfaces.Strategy(&mixedStrategy1), nil
	default:
		return nil, fmt.Errorf("%s is not a known strategy", strategyName)
	}

}
