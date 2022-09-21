package strategies

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/aoterocom/AOCryptobot/helpers"
	"gitlab.com/aoterocom/AOCryptobot/strategies"
	"gitlab.com/aoterocom/AOCryptobot/tests/mocks"
	"testing"
)

func TestLun1MarCustomStrategy(t *testing.T) {
	interval := "1h"
	strategy := strategies.NewLun1MarCustomStrategy(interval)
	exchangeService := &mocks.ProviderMock{}
	symbol := "ETHEUR"
	strategyResults, _ := strategy.PerformSimulation(symbol, exchangeService, interval, 240, 0, nil)

	ratio := helpers.PositiveNegativeRatio(strategyResults.ProfitList)
	profit := strategyResults.Profit
	constants := strategyResults.Constants

	fmt.Printf("%s: Best combination last 500 candles: ", symbol)
	for i, _ := range strategyResults.Constants {
		fmt.Printf("Constant%d %.8f ", i+1, strategyResults.Constants[i])
	}
	fmt.Printf("Ratio %f Profit: %.2f%%\n", helpers.PositiveNegativeRatio(strategyResults.ProfitList),
		strategyResults.Profit)

	assert.Equal(t, 1.0, ratio)
	assert.Equal(t, 3.1977614488366584, profit)
	assert.Equal(t, []float64{0.07200000000000001}, constants)
}
