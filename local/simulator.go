package main

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	database "gitlab.com/aoterocom/AOCryptobot/database"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models"
	binance "gitlab.com/aoterocom/AOCryptobot/providers/paper"
	"gitlab.com/aoterocom/AOCryptobot/strategies"
	"log"
	"os"
	"time"
)

func main() {

	dbService := database.NewDBService("AOCryptoBot", "database-1.cxb3uopd6ntp.eu-west-3.rds.amazonaws.com", "3306", "admin", "abc123..")

	start, _ := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Feb 4, 2014 at 6:05pm (PST)")
	end, _ := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Feb 5, 2014 at 6:05pm (PST)")

	candle := techan.Candle{
		Period: techan.TimePeriod{
			Start: start,
			End:   end,
		},
		OpenPrice:  big.NewDecimal(0.1),
		ClosePrice: big.NewDecimal(0.1),
		MaxPrice:   big.NewDecimal(0.1),
		MinPrice:   big.NewDecimal(0.1),
		Volume:     big.NewDecimal(0.1),
		TradeCount: 10000,
	}

	dbService.AddOrUpdateCandle(candle, "NEWEUR")

	//
	//candle = techan.Candle{
	//	Period:     techan.TimePeriod{
	//		Start: start,
	//		End:   end,
	//	},
	//	OpenPrice:  big.NewDecimal(3.1),
	//	ClosePrice: big.NewDecimal(3.1),
	//	MaxPrice:   big.NewDecimal(3.1),
	//	MinPrice:   big.NewDecimal(3.1),
	//	Volume:     big.NewDecimal(3.1),
	//	TradeCount: 10000,
	//}
	//
	//dbService.AddOrUpdateCandle(candle, "NEW2EUR")

	os.Exit(1)

	dbOrder1 := models.NewEmptyOrder()
	dbOrder1.StopPrice = "SISISI"
	dbPosition := models.NewPosition(dbOrder1)

	id := dbService.AddPosition(*dbPosition, "NEW", []float64{3.0, 2.1}, 0.0, 0.0, 0.0)

	dbOrder2 := models.NewEmptyOrder()
	dbOrder2.StopPrice = "TOMATOMATOMA"

	dbPosition.Exit(dbOrder2)
	dbService.UpdatePosition(id, *dbPosition, "NEW", []float64{3.0, 2.1}, 0.0, 0.0, 0.0)

	os.Exit(1)
	targetCoin := "EUR"
	var whitelist []string
	var blacklist []string
	bs := binance.NewPaperService()
	exchangeService := interfaces.ExchangeService(bs)

	//lun1MarCustomStrategy := strategies.NewLun1MarCustomStrategy()
	//lun5JulCustomStrategy := strategies.NewLun5JulCustomStrategy()
	//MACDCustomStrategy := strategies.NewMACDCustomStrategy()
	//stableStrategy := strategies.NewStableStrategy()
	//stochRSICustomStrategy := strategies.NewStochRSICustomStrategy()
	mixedStrategy1 := strategies.NewMixedStrategy1("1h")

	selectedStrategies := []interfaces.Strategy{
		//&MACDCustomStrategy,
		//&lun1MarCustomStrategy,
		//&stochRSICustomStrategy,
		&mixedStrategy1,
		//&lun5JulCustomStrategy,
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
