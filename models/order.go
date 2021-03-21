package models

type Order struct {
	Symbol                  string          `json:"symbol"`
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

func NewOrder(symbol string, orderID int64, clientOrderID string, price string, origQuantity string, executedQuantity string, cumulativeQuoteQuantity string, status OrderStatusType, type_ OrderType, side SideType, time int64, updateTime int64, isWorking bool, isIsolated bool) Order {
	return Order{
		Symbol:                  symbol,
		OrderID:                 orderID,
		ClientOrderID:           clientOrderID,
		Price:                   price,
		OrigQuantity:            origQuantity,
		ExecutedQuantity:        executedQuantity,
		CumulativeQuoteQuantity: cumulativeQuoteQuantity,
		Status:                  status,
		Type:                    type_,
		Side:                    side,
		Time:                    time,
		UpdateTime:              updateTime,
		IsWorking:               isWorking,
		IsIsolated:              isIsolated,
	}
}

func NewEmptyOrder() Order {
	return Order{}
}

// OrderSide is a simple enumeration representing the side of an Order (buy or sell)
type OrderSide int

// BUY and SELL enumerations
const (
	BUY OrderSide = iota
	SELL
)

// OrderStatusType define order status type
type OrderStatusType string

// OrderType define order type
type OrderType string

// SideType define side type
type SideType string

// TradeDirection define the direction of a position
type MarketDirection string

// Global enums
const (
	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"

	MarketDirectionLong  MarketDirection = "BUY"
	MarketDirectionShort MarketDirection = "SELL"

	OrderTypeLimit           OrderType = "LIMIT"
	OrderTypeMarket          OrderType = "MARKET"
	OrderTypeLimitMaker      OrderType = "LIMIT_MAKER"
	OrderTypeStopLoss        OrderType = "STOP_LOSS"
	OrderTypeStopLossLimit   OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"

	OrderStatusTypeNew             OrderStatusType = "NEW"
	OrderStatusTypePartiallyFilled OrderStatusType = "PARTIALLY_FILLED"
	OrderStatusTypeFilled          OrderStatusType = "FILLED"
	OrderStatusTypeCanceled        OrderStatusType = "CANCELED"
	OrderStatusTypePendingCancel   OrderStatusType = "PENDING_CANCEL"
	OrderStatusTypeRejected        OrderStatusType = "REJECTED"
	OrderStatusTypeExpired         OrderStatusType = "EXPIRED"
)
