package interfaces

import "github.com/sdcoffey/techan"

type Strategy interface {
	ShouldEnter(timeSeries *techan.TimeSeries) bool
	ShouldExit(timeSeries *techan.TimeSeries) bool
}
