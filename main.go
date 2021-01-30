package main

import (
	"./marketmaker"
	"./marketmaker/service/binance"
	"os"
)

func main() {
	marketService := service.MarketService{}
	marketService.SetPair(os.Getenv("pair"))
	marketService.ConfigureClient()
	marketService.StartMonitor()
	run(&marketService)
}

func run(marketService *service.MarketService) {
	strategy := marketmaker.MMStrategy{}
	strategy.SetMarketService(marketService)
	strategy.Execute()
}
