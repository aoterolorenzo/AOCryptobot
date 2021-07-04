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
	bs := binance.NewPaperService()
	exchangeService := interfaces.ExchangeService(bs)

	lun1MarCustomStrategy := strategies.NewLun1MarCustomStrategy()
	MACDCustomStrategy := strategies.NewMACDCustomStrategy()
	stableStrategy := strategies.NewStableStrategy()
	stochRSICustomStrategy := strategies.NewStochRSICustomStrategy()

	selectedStrategies := []interfaces.Strategy{
		&lun1MarCustomStrategy,
		&stochRSICustomStrategy,
		&MACDCustomStrategy,
		&stableStrategy,
	}

	for _, pair := range exchangeService.GetMarkets(targetCoin) {
		helpers.Logger.Infoln("⚠️ Analyzing " + pair + " ⚠️")
		for _, strategy := range selectedStrategies {
			_, err := strategy.Analyze(pair, exchangeService)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

}
