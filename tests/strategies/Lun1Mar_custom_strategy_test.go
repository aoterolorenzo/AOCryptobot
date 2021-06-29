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
	strategy := strategies.NewLun1MarCustomStrategy()
	exchangeService := &mocks.ProviderMock{}
	symbol := "ETHEUR"
	strategyResults, _ := strategy.PerformSimulation(symbol, exchangeService, "1h", 500, 0, nil)

	ratio := helpers.PositiveNegativeRatio(strategyResults.ProfitList)
	profit := strategyResults.Profit
	constants := strategyResults.Constants

	fmt.Printf("%s: Best combination last 500 candles: ", symbol)
	for i, _ := range strategyResults.Constants {
		fmt.Printf("Constant%d %.8f ", i+1, strategyResults.Constants[i])
	}
	fmt.Printf("Ratio %f Profit: %.2f%%\n", helpers.PositiveNegativeRatio(strategyResults.ProfitList),
		strategyResults.Profit)

	assert.Equal(t, 0.7857142857142857, ratio)
	assert.Equal(t, 6.953072770110964, profit)
	assert.Equal(t, []float64{0.005}, constants)
}
