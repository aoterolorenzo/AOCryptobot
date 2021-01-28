##AOCryptobot

###What is this

AOCryptobot is a simple (and in an early development, almost scripting) Market Maker strategy bot which tries to get 
profit from the oscilations and the spread between the bids and the asks of a trading pair.

A bot... and a Go learning experience for myself.

###Current development

This is currently something between a meme and a frankenstein. The bot was develop in just a couple of days,
and it has some meme workarounds and bugs. 

Firstable, it endless print to stdout announcing the status of the market

It uses 2 different go binance libraries, the original one, and a new one with the ability to make OCO market orders.
The original one ("github.com/eranyanay/binance-api") being used to get the market status each time period, so the
change to a completely new one implied a full refactor of the bot, and I hadn't the necessary amount of time to do that,
so I'm using the new one ("github.com/adshao/go-binance/v2") to make the OCO offers.

I also had some troubles trying to generate multiple threads because of the binance API rate limits, so I intend to do
a full refactor with the new library and use a binance websocket subscription in a future, final deprecating the
original one and avoiding that TOC that produces me seeing "binance" and "binance2" variables.

Get the output properly treated is also important, but as I said, this was a time trial to me on the correction 
(crypto bubble) formed each year in christmas and january.

Also support for other exchanges is enabled, and a binance-paper service is something that I did but not commit to this
project due it's poor implementation.


###Behaviour


####Market monitor
The bot monitors the market during `monitorWindow` seconds with a `monitorFrequency` also in seconds.

####Buy
The signal it uses to do a buy is calculated through the percentil in which the current center price is, and
a strategy-given variables `topPercentileToTrade` and `lowPercentileToTrade`. This means that if the current center
price is on a percentil between that values, the buy is triggered.

Example:
Monitor window higher price: 1100
Monitor window lower price: 900
topPercentileToTrade: 100 
lowPercentileToTrade: 50

The buy will be triggered when the current center price be from 1000 and 1000.
The amount and the value of the buy will be defined in the strategy (`strategy.env`)
as `pctAmountToTrade` (over your wallet total) and `buyMargin` (pct below the center price)

####Sell
At the moment that a buy is successful, a OCO sell order is triggered.

An OCO order is a One-Cancel-Other order, so it implies the creation or 2 mutually exclusive order:
 - Sell order for the value of the buy including the `sellMargin` (pct over the buy cost)
 - Stop loss-limit order for the value of the `stopLossPct` percentage
 
That implies that the sell order will be complete or taking the benefit or by the stop-loss percentage

#####This buy-sell process will be infinite repeated
 


###Usage


1. Install dependencies with
```go get ./..```

2. Test with
```go test ./..```

2. Run with
```go run main.go```


###Strategy file

The strategy file is `marketmaker/strategy.env`

###Open Source

The development is open to everyone and everything, to new market strategies such as arbitrage,
perpetual market making... Just use as you need.

I would like to integrate market signals as MACD, RSI, BoilerBands, etcetc... and create a real strategy, being able 
to choose between or combine them.

A very interesting lib to do it is github.com/sdcoffey/techan