package main

import (
	"gitlab.com/aoterocom/AOCryptobot/bot_oscillator"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
)

func main() {
	bot := &bot_oscillator.MarketMaker{}
	rubBot(bot)
}

func rubBot(b interfaces.MarketBot) {
	b.Run()
}
