package models

import (
	"strconv"
)

// Position is a pair of two Order objects
type Position struct {
	orders [2]*Order
}

// NewPosition returns a new Position with the passed-in order as the open order
func NewPosition(openOrder Order) (t *Position) {
	t = new(Position)
	t.orders[0] = &openOrder

	return t
}

// Enter sets the open order to the order passed in
func (p *Position) Enter(order Order) {
	p.orders[0] = &order
}

// Exit sets the exit order to the order passed in
func (p *Position) Exit(order Order) {
	p.orders[1] = &order
}

// IsLong returns true if the entrance order is a buy order
func (p *Position) IsLong() bool {
	return p.EntranceOrder() != nil && p.EntranceOrder().Side == SideTypeBuy
}

// IsShort returns true if the entrance order is a sell order
func (p *Position) IsShort() bool {
	return p.EntranceOrder() != nil && p.EntranceOrder().Side == SideTypeSell
}

// IsOpen returns true if there is an entrance order but no exit order
func (p *Position) IsOpen() bool {
	return p.EntranceOrder() != nil && p.ExitOrder() == nil
}

// IsClosed returns true of there are both entrance and exit orders
func (p *Position) IsClosed() bool {
	return p.EntranceOrder() != nil && p.ExitOrder() != nil
}

// IsNew returns true if there is neither an entrance or exit order
func (p *Position) IsNew() bool {
	return p.EntranceOrder() == nil && p.ExitOrder() == nil
}

// EntranceOrder returns the entrance order of this position
func (p *Position) EntranceOrder() *Order {
	return p.orders[0]
}

// ExitOrder returns the exit order of this position
func (p *Position) ExitOrder() *Order {
	return p.orders[1]
}

// CostBasis returns the price to enter this order
func (p *Position) CostBasis() float64 {
	if p.EntranceOrder() != nil {
		cumulativeQuoteQuantity, _ := strconv.ParseFloat(p.EntranceOrder().CumulativeQuoteQuantity, 64)
		return cumulativeQuoteQuantity
	}
	return -1.0
}

// ExitValue returns the value accrued by closing the position
func (p *Position) ExitValue() float64 {
	if p.IsClosed() {
		executedQuantity, _ := strconv.ParseFloat(p.ExitOrder().ExecutedQuantity, 64)
		price, _ := strconv.ParseFloat(p.ExitOrder().Price, 64)
		return executedQuantity * price
	}
	return -1.0
}

// Profit returns the current/final profit coin amount
func (p *Position) ProfitPct() float64 {
	if p.IsOpen() {
		return -1.0
	}
	entranceOrder := p.EntranceOrder()
	exitOrder := p.ExitOrder()
	enterCumulativeQuoteQuantity, _ := strconv.ParseFloat(entranceOrder.CumulativeQuoteQuantity, 64)
	exitCumulativeQuoteQuantity, _ := strconv.ParseFloat(exitOrder.CumulativeQuoteQuantity, 64)
	return (exitCumulativeQuoteQuantity/enterCumulativeQuoteQuantity - 1) * (1 - 0.0015)
}
