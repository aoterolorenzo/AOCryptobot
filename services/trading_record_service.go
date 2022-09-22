package services

import (
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TradingRecordService struct {
	OpenPositions        map[string][]*models.Position
	ClosedPositions      map[string][]*models.Position
	openPositionsMutex   *sync.Mutex
	closedPositionsMutex *sync.Mutex

	multiMarketService *MultiMarketService
	exchangeService    interfaces.ExchangeService
}

func NewTradingRecordService(multiMarketService *MultiMarketService, exchangeService interfaces.ExchangeService) *TradingRecordService {
	return &TradingRecordService{
		OpenPositions:        make(map[string][]*models.Position),
		ClosedPositions:      make(map[string][]*models.Position),
		openPositionsMutex:   &sync.Mutex{},
		closedPositionsMutex: &sync.Mutex{},
		multiMarketService:   multiMarketService,
		exchangeService:      exchangeService,
	}
}

func (trs *TradingRecordService) AddPosition(position models.Position) error {
	trs.openPositionsMutex.Lock()
	trs.OpenPositions[position.EntranceOrder().Symbol] = append(trs.OpenPositions[position.EntranceOrder().Symbol], &position)
	trs.openPositionsMutex.Unlock()
	return nil
}

func (trs *TradingRecordService) EnterPosition(pair string, amount float64, direction models.MarketDirection) error {

	var orderSide models.OrderSide
	if direction == models.MarketDirectionLong {
		orderSide = models.BUY
	} else {
		orderSide = models.SELL
	}

	currentRate := trs.multiMarketService.GetTimeSeries(pair).LastCandle().ClosePrice
	order, err := trs.exchangeService.MakeOrder(pair, amount, currentRate.Float(), models.OrderTypeMarket, orderSide)
	if err != nil {
		helpers.Logger.Errorln(err)
		return err
	}
	position := models.NewPosition(order)
	trs.GrabMemoryPosition(pair, position)
	return nil
}

func (trs *TradingRecordService) GrabMemoryPosition(pair string, position *models.Position) {
	trs.openPositionsMutex.Lock()
	trs.OpenPositions[pair] = append(trs.OpenPositions[position.EntranceOrder().Symbol], position)
	trs.openPositionsMutex.Unlock()
}

func (trs *TradingRecordService) ExitPositions(pair string, direction models.MarketDirection) error {

	var orderSide models.OrderSide
	if direction == models.MarketDirectionLong {
		orderSide = models.SELL
	} else {
		orderSide = models.BUY
	}

	currentRate := trs.multiMarketService.GetTimeSeries(pair).LastCandle().ClosePrice
	trs.openPositionsMutex.Lock()
	pairOpenPositions := trs.OpenPositions[pair]
	trs.openPositionsMutex.Unlock()
	for i, openPosition := range pairOpenPositions {

		var quantity float64
		if direction == models.MarketDirectionLong {
			quantity, _ = strconv.ParseFloat(openPosition.EntranceOrder().ExecutedQuantity, 64)
		} else {
			quantity, _ = strconv.ParseFloat(openPosition.EntranceOrder().CumulativeQuoteQuantity, 64)
		}

		order, err := trs.exchangeService.MakeOrder(pair, quantity, currentRate.Float(), models.OrderTypeMarket, orderSide)
		if err != nil {
			count := 0
			helpers.Logger.Errorln(err)
			for strings.Contains(err.Error(), "Account has insufficient balance for requested action") {
				quantity *= 1 - 0.0005
				order, err = trs.exchangeService.MakeOrder(pair, quantity, currentRate.Float(), models.OrderTypeMarket, orderSide)
				if err == nil {
					break
				} else if count > 15 {
					helpers.Logger.Fatalln(err)
					return err
				}

				count++
				time.Sleep(500 * time.Millisecond)
			}
		}
		openPosition.Exit(order)
		trs.closedPositionsMutex.Lock()
		trs.ClosedPositions[pair] = append(trs.ClosedPositions[pair], openPosition)
		trs.closedPositionsMutex.Unlock()
		trs.openPositionsMutex.Lock()
		trs.OpenPositions[pair] = append(trs.OpenPositions[pair][:i], trs.OpenPositions[pair][i+1:]...)
		trs.openPositionsMutex.Unlock()
	}
	return nil
}

func (trs *TradingRecordService) OpenPositionsCount() int {
	count := 0
	trs.openPositionsMutex.Lock()
	openPositions := trs.OpenPositions
	for _, openPairPositions := range openPositions {
		count += len(openPairPositions)
	}
	trs.openPositionsMutex.Unlock()
	return count
}

func (trs *TradingRecordService) ClosedPositionsCount() int {
	count := 0
	for _, closedPairPositions := range trs.ClosedPositions {
		count += len(closedPairPositions)
	}
	return count
}

func (trs *TradingRecordService) HasOpenPositions(pair string) bool {
	trs.openPositionsMutex.Lock()
	hasOpenPositions := len(trs.OpenPositions[pair]) > 0
	trs.openPositionsMutex.Unlock()
	return hasOpenPositions
}

func (trs *TradingRecordService) LastClosedPosition(pair string) *models.Position {
	trs.closedPositionsMutex.Lock()
	position := trs.ClosedPositions[pair][len(trs.ClosedPositions[pair])-1]
	trs.closedPositionsMutex.Unlock()
	return position
}

func (trs *TradingRecordService) LastOpenPosition(pair string) *models.Position {
	trs.closedPositionsMutex.Lock()
	position := trs.OpenPositions[pair][len(trs.OpenPositions[pair])-1]
	trs.closedPositionsMutex.Unlock()
	return position
}
