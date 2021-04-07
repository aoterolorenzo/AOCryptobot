package helpers

import "math"

func PositiveNegativeRatio(list []float64) float64 {
	countPositive := 0
	countNegative := 0
	for _, item := range list {
		if item > 0 {
			countPositive++
		} else {
			countNegative++
		}
	}

	if countNegative == 0 {
		return 0
	}
	return float64(countPositive) / float64(countNegative)
}

func StdDev(numbers []float64, mean float64) float64 {
	total := 0.0
	for _, number := range numbers {
		total += math.Pow(number-mean, 2)
	}
	variance := total / float64(len(numbers)-1)
	return math.Sqrt(variance)
}

func Sum(numbers []float64) (total float64) {
	for _, x := range numbers {
		total += x
	}
	return total
}

func AllValuesPositive(list []float64) bool {
	for _, item := range list {
		if item < 0.0 {
			return false
		}
	}
	return true
}
