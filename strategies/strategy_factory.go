package strategies

import (
	"fmt"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
)

func StrategyFactory(strategyName string) (interfaces.Strategy, error) {

	switch strategyName {
	case "lun1MarCustomStrategy":
		lun1MarCustomStrategy := NewLun1MarCustomStrategy()
		return interfaces.Strategy(&lun1MarCustomStrategy), nil
	case "lun5JulCustomStrategy":
		lun5JulCustomStrategy := NewLun5JulCustomStrategy()
		return interfaces.Strategy(&lun5JulCustomStrategy), nil
	case "stableStrategy":
		stableStrategy := NewStableStrategy()
		return interfaces.Strategy(&stableStrategy), nil
	case "MACDCustomStrategy":
		MACDCustomStrategy := NewMACDCustomStrategy()
		return interfaces.Strategy(&MACDCustomStrategy), nil
	case "stochRSICustomStrategy":
		stochRSICustomStrategy := NewStochRSICustomStrategy()
		return interfaces.Strategy(&stochRSICustomStrategy), nil
	case "MACDInStochRSIOutCustomStrategy":
		MACDInStochRSIOutCustomStrategy := NewMACDInStochRSIOutCustomStrategy()
		return interfaces.Strategy(&MACDInStochRSIOutCustomStrategy), nil
	case "stochRSIInMACDOutCustomStrategy":
		stochRSIInMACDOutCustomStrategy := NewStochRSIInMACDOutCustomStrategy()
		return interfaces.Strategy(&stochRSIInMACDOutCustomStrategy), nil
	case "mixedStrategy1":
		mixedStrategy1 := NewMixedStrategy1()
		return interfaces.Strategy(&mixedStrategy1), nil
	default:
		return nil, fmt.Errorf("%s is not a known strategy", strategyName)
	}

}
