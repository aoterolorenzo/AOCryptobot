package main

import (
	botSignalTrader "gitlab.com/aoterocom/AOCryptobot/bot_signal-trader"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
)

func main() {
	bot := &botSignalTrader.SignalTrader{}
	rubBot(bot)
}

func rubBot(b interfaces.MarketBot) {
	b.Run()
}
