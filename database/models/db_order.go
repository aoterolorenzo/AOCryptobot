package database

import "gorm.io/gorm"

// OrderStatusType define order status type
type OrderStatusType string

// OrderType define order type
type OrderType string

// SideType define side type
type SideType string

// TradeDirection define the direction of a position
type MarketDirection string

type Order struct {
	gorm.Model
	PositionID              uint
	OrderID                 int64           `json:"orderId"`
	ClientOrderID           string          `json:"clientOrderId"`
	Price                   string          `json:"price"`
	OrigQuantity            string          `json:"origQty"`
	ExecutedQuantity        string          `json:"executedQty"`
	CumulativeQuoteQuantity string          `json:"cumulativeQuoteQty"`
	Status                  OrderStatusType `json:"status"`
	Type                    OrderType       `json:"type"`
	Side                    SideType        `json:"side"`
	StopPrice               string          `json:"stopPrice"`
	IcebergQuantity         string          `json:"icebergQty"`
	Time                    int64           `json:"time"`
	UpdateTime              int64           `json:"updateTime"`
	IsWorking               bool            `json:"isWorking"`
	IsIsolated              bool            `json:"isIsolated"`
}
