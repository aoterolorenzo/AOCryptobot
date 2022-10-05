package helpers

import (
	"github.com/xhit/go-str2duration/v2"
	"math"
	"strconv"
	"strings"
)

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

func StringIntervalToSeconds(interval string) float64 {
	interval = strings.ReplaceAll(interval, "S", "w")
	interval = strings.ToLower(interval)
	durationFromString, err := str2duration.ParseDuration(interval)
	if err != nil {
		return -1.0
	}
	return durationFromString.Seconds()
}

func RemoveFromSlice(slice []float64, s int) []float64 {
	return append(slice[:s], slice[s+1:]...)
}

func IntervalFromString(str string) int {

	var count int
	if strings.HasSuffix(str, "s") {
		count, _ = strconv.Atoi(strings.Replace(str, "s", "", 1))
		return count
	} else if strings.HasSuffix(str, "m") {
		count, _ = strconv.Atoi(strings.Replace(str, "m", "", 1))
		return count * 60
	} else if strings.HasSuffix(str, "h") {
		count, _ = strconv.Atoi(strings.Replace(str, "h", "", 1))
		return count * 60 * 60
	} else if strings.HasSuffix(str, "D") {
		count, _ = strconv.Atoi(strings.Replace(str, "D", "", 1))
		return count * 60 * 60 * 24
	} else if strings.HasSuffix(str, "S") {
		count, _ = strconv.Atoi(strings.Replace(str, "S", "", 1))
		return count * 60 * 60 * 24 * 7
	} else if strings.HasSuffix(str, "M") {
		count, _ = strconv.Atoi(strings.Replace(str, "M", "", 1))
		return count * 60 * 60 * 24 * 30
	}

	return 0
}
