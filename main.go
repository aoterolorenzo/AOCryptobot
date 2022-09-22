package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	botSignalTrader "gitlab.com/aoterocom/AOCryptobot/bot_signal-trader"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"log"
	"os"
	"time"
)

func main() {

	app := &cli.App{
		Name:  "AOCryptoBot",
		Usage: "Advanced crypto trading bot",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "strategies",
				Aliases: []string{"s"},
				Usage:   "strategies list, comma separated",
			},
		},
		Action: func(c *cli.Context) error {
			bot := &botSignalTrader.SignalTrader{}
			rubBot(bot, c)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func rubBot(b interfaces.MarketBot, c *cli.Context) {
	// Panic recovery. Relaunch application.
	defer func() {
		if r := recover(); r != nil {
			helpers.Logger.Errorln(fmt.Sprintf("Recovered. Error:\n %v", r))
			time.Sleep(1 * time.Second)
			rubBot(b, c)
		}
	}()
	b.Run(c)
}
