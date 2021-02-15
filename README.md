## AOCryptobot

### Current development

AOCryptobot is a simple (and in an early development, almost scripting) Market Maker strategy bot which tries to get 
profit from the oscilations and the spread between the bids and the asks of a trading pair.

A bot... and a Go learning experience for myself.

![](./images/screenshot.png)

### Changelog

The initial development was made in about a couple of days. 

[09 FEB 2020]
- **Add user interface**: A graphic interface has been integrated. Still some bugs to make it stable, but milestone is not so far away.
- Fix order cancel and get sellingTimeout correctly working
- Resolve "Refactor binance monitor replacing REST with websocket calls"
- Update conf.env
- Add CODEOWNERS file
### Behaviour


#### Market monitor
The bot monitors the market during `monitorWindow` seconds with a `monitorFrequency` also in seconds.

#### Buy
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

#### Sell
At the moment that a buy is successful, a OCO sell order is triggered.

An OCO order is a One-Cancel-Other order, so it implies the creation or 2 mutually exclusive order:
 - Sell order for the value of the buy including the `sellMargin` (pct over the buy cost)
 - Stop loss-limit order for the value of the `stopLossPct` percentage
 
That implies that the sell order will be complete or taking the benefit or by the stop-loss percentage

##### This buy-sell process will be infinite repeated
 


### Usage


1. Install dependencies with
`go mod download`

2. Test with
`go test ./..`

2. Run with
`go run main.go`


### Strategy file

The strategy file is `marketmaker/strategy.env`

### Open Source

The development is open to everyone and everything, to new market strategies such as arbitrage,
perpetual market making... Just use as you need.

I would like to integrate market signals as MACD, RSI, BoilerBands, etcetc... and create a real strategy, being able 
to choose between or combine them.

A very interesting lib to do it is github.com/sdcoffey/techan
