package main

import (
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	binance "gitlab.com/aoterocom/AOCryptobot/providers/paper"
	"gitlab.com/aoterocom/AOCryptobot/strategies"
	"log"
)

func main() {

	targetCoin := "EUR"
	var whitelist []string
	var blacklist []string
	bs := binance.NewPaperService()
	exchangeService := interfaces.ExchangeService(bs)

	lun1MarCustomStrategy := strategies.NewLun1MarCustomStrategy("1h")
	lun5JulCustomStrategy := strategies.NewLun5JulCustomStrategy("1h")
	MACDCustomStrategy := strategies.NewMACDCustomStrategy("1h")
	//stableStrategy := strategies.NewStableStrategy()
	stochRSICustomStrategy := strategies.NewStochRSICustomStrategy("1h")
	mixedStrategy1 := strategies.NewMixedStrategy1("1h")

	selectedStrategies := []interfaces.Strategy{
		&MACDCustomStrategy,
		&lun1MarCustomStrategy,
		&stochRSICustomStrategy,
		&mixedStrategy1,
		&lun5JulCustomStrategy,
		//&stableStrategy,
	}

	for _, pair := range exchangeService.GetMarkets(targetCoin, whitelist, blacklist) {
		helpers.Logger.Infoln("⚠️ Analyzing " + pair + " ⚠️")
		for _, strategy := range selectedStrategies {
			_, err := strategy.Analyze(pair, exchangeService)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

}
