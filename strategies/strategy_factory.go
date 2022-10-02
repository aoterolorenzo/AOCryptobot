package strategies

import (
	"fmt"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
)

type STString map[string]func(interval string)

func STStrings() map[string]func(interval string) interfaces.Strategy {
	strategyMap := make(map[string]func(interval string) interfaces.Strategy)

	strategyMap["lun1MarCustomStrategy"] = func(interval string) interfaces.Strategy {
		strategy := NewLun1MarCustomStrategy(interval)
		return &strategy
	}

	strategyMap["lun5JulCustomStrategy"] = func(interval string) interfaces.Strategy {
		strategy := NewLun5JulCustomStrategy(interval)
		return &strategy
	}

	strategyMap["stableStrategy"] = func(interval string) interfaces.Strategy {
		strategy := NewStableStrategy(interval)
		return &strategy
	}

	strategyMap["MACDCustomStrategy"] = func(interval string) interfaces.Strategy {
		strategy := NewMACDCustomStrategy(interval)
		return &strategy
	}

	strategyMap["stochRSICustomStrategy"] = func(interval string) interfaces.Strategy {
		strategy := NewStochRSICustomStrategy(interval)
		return &strategy
	}

	strategyMap["MACDInStochRSIOutCustomStrategy"] = func(interval string) interfaces.Strategy {
		strategy := NewMACDInStochRSIOutCustomStrategy(interval)
		return &strategy
	}

	strategyMap["stochRSIInMACDOutCustomStrategy"] = func(interval string) interfaces.Strategy {
		strategy := NewStochRSIInMACDOutCustomStrategy(interval)
		return &strategy
	}

	strategyMap["mixedStrategy1"] = func(interval string) interfaces.Strategy {
		strategy := NewMixedStrategy1(interval)
		return &strategy
	}

	strategyMap["stopLossTriggerStrategy"] = func(interval string) interfaces.Strategy {
		strategy := StopLossTriggerStrategy{}
		return &strategy
	}

	return strategyMap
}

func StrategyFactory(strategyName string, interval string) (interfaces.Strategy, error) {
	for strategyString, constructor := range STStrings() {
		if strategyName == strategyString {
			return constructor(interval), nil
		}
	}

	return nil, fmt.Errorf("%s is not a known strategy", strategyName)
}
