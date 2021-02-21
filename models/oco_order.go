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
