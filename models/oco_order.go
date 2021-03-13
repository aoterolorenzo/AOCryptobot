package models

type OCOOrder struct {
	OrderListID       int64   `json:"orderListId"`
	ContingencyType   string  `json:"contingencyType"`
	ListStatusType    string  `json:"listStatusType"`
	ListOrderStatus   string  `json:"listOrderStatus"`
	ListClientOrderID string  `json:"listClientOrderId"`
	TransactionTime   int64   `json:"transactionTime"`
	Symbol            string  `json:"symbol"`
	Orders            []Order `json:"orders"`
}

func NewOCOOrder(orderListID int64, contingencyType string, listStatusType string, listOrderStatus string,
	listClientOrderID string, transactionTime int64, symbol string) OCOOrder {
	return OCOOrder{
		OrderListID:       orderListID,
		ContingencyType:   contingencyType,
		ListStatusType:    listStatusType,
		ListOrderStatus:   listOrderStatus,
		ListClientOrderID: listClientOrderID,
		TransactionTime:   transactionTime,
		Symbol:            symbol,
	}
}
