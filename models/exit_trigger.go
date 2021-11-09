package models

type ExitTrigger string

const (
	ExitTriggerStopLoss         ExitTrigger = "Stop Loss"
	ExitTriggerTrailingStopLoss             = "Trailing Stop Loss"
	ExitTriggerStrategy                     = "Strategy"
	ExitTriggerNone                         = ""
)
