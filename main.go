package main

import (
	"gitlab.com/aoterocom/AOCryptobot/bot_oscillator"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/providers/binance"
)

func main() {
	bot := &bot_oscillator.MarketMaker{}
	rubBot(bot)

	ps := &binance.BinanceService{}
	rP(ps)
}

func rubBot(b interfaces.MarketBot) {
	b.Run()
}

func rP(b interfaces.ExchangeService) {
	b.ConfigureClient()
}
