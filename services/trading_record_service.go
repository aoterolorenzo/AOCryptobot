package services

import (
	"gitlab.com/aoterocom/AOCryptobot/interfaces"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"strconv"
)

type TradingRecordService struct {
	OpenPositions   map[string][]*models.Position
	ClosedPositions map[string][]*models.Position
	exchangeService interfaces.ExchangeService
}

func NewTradingRecordService(marketService interfaces.ExchangeService) *TradingRecordService {
	return &TradingRecordService{
		OpenPositions:   make(map[string][]*models.Position),
		ClosedPositions: make(map[string][]*models.Position),
		exchangeService: marketService,
	}
}

func (trs *TradingRecordService) AddPosition(position models.Position) error {
	trs.OpenPositions[position.EntranceOrder().Symbol] = append(trs.OpenPositions[position.EntranceOrder().Symbol], &position)
	return nil
}

func (trs *TradingRecordService) EnterPosition(pair string, amount float64) error {
	order, err := trs.exchangeService.MakeOrder(pair, amount, 0, models.OrderTypeMarket, models.BUY)
	if err != nil {
		return err
	}
	position := models.NewPosition(order)
	trs.OpenPositions[pair] = append(trs.OpenPositions[position.EntranceOrder().Symbol], position)
	return nil
}

func (trs *TradingRecordService) ExitPositions(pair string) error {
	for i, openPosition := range trs.OpenPositions[pair] {
		executedQuantity, err := strconv.ParseFloat(openPosition.EntranceOrder().ExecutedQuantity, 64)
		order, err := trs.exchangeService.MakeOrder(pair, executedQuantity, 0, models.OrderTypeMarket, models.SELL)
		if err != nil {
			return err
		}
		openPosition.Exit(order)
		trs.ClosedPositions[pair] = append(trs.ClosedPositions[pair], openPosition)
		trs.OpenPositions[pair] = append(trs.OpenPositions[pair][:i], trs.OpenPositions[pair][i+1:]...)
	}
	return nil
}

func (trs *TradingRecordService) OpenPositionsCount() int {
	count := 0
	for _, openPairPositions := range trs.OpenPositions {
		count += len(openPairPositions)
	}
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
	return len(trs.OpenPositions[pair]) > 0
}
