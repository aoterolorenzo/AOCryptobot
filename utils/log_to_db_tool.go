package main

import (
	"bufio"
	"fmt"
	database "gitlab.com/aoterocom/AOCryptobot/database/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	gorm.Model
	OpenPositions map[string][]*Position
}

type Position struct {
	gorm.Model
	Orders []Order `gorm:"foreignKey:Position"`
}

type Order struct {
	gorm.Model
	Position uint
	Symbol   string
	OrderID  int64
	Price    string
}

func main() {
	file, err := os.Open("bot.log")

	if err != nil {
		panic("al abrir archivo")
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	dsn := "user:pass@tcp(127.0.0.1:3306)/AOCryptoBot?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	_ = db.AutoMigrate(&database.Position{})
	_ = db.AutoMigrate(&database.Order{})
	_ = db.AutoMigrate(&database.Constant{})

	timeLayout := "2006-01-02 15:04:05"

	openOrders := make(map[string]database.Order)
	currentLine := ""
	for {
		currentLine, err = reader.ReadString('\n')
		for !strings.Contains(currentLine, " signal**") {
			currentLine, err = reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
		}

		if strings.Contains(currentLine, "Entry") {
			fmt.Println("Entry signal")
			r := regexp.MustCompile("INFO  ([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}) 📈 \\*\\*([A-Z]+)")
			res := r.FindStringSubmatch(currentLine)
			symbol := res[2]
			date := res[1]
			fmt.Println(symbol)
			fmt.Println(date)

			t, err := time.Parse(timeLayout, res[1])
			if err != nil {
				panic("Error parsing time")
			}

			line2, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Strategy: (.*)")
			strategy := r.FindStringSubmatch(line2)[1]
			fmt.Println(strategy)

			line3, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Constants: \\[([0-9]+\\.[0-9]+)\\s?([0-9]+\\.[0-9]+)?\\s?([0-9]+\\.[0-9]+)?\\]")
			constantsString := r.FindStringSubmatch(line3)

			var constants []float64
			for i := 1; i < len(constantsString); i++ {
				if constantsString[i] != "" {
					num, _ := strconv.ParseFloat(constantsString[i], 64)
					constants = append(constants, num)
				}
			}
			fmt.Println(constants)

			line4, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Buy Price: (.*)")
			price := r.FindStringSubmatch(line4)[1]
			fmt.Println(price)

			_, _ = reader.ReadString('\n')

			line6, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Updated currentBalance: (.*)")
			balance, _ := strconv.ParseFloat(r.FindStringSubmatch(line6)[1], 64)
			balance = balance / 10
			fmt.Println(balance)

			openOrders[symbol] = database.Order{
				Price:                   price,
				Time:                    t.Unix(),
				ExecutedQuantity:        fmt.Sprintf("%f", balance),
				CumulativeQuoteQuantity: fmt.Sprintf("%f", balance),
				OrigQuantity:            fmt.Sprintf("%f", balance),
			}

		} else {
			fmt.Println("Exit signal")
			r := regexp.MustCompile("INFO  ([0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}) 📉 \\*\\*([A-Z]+)")
			res := r.FindStringSubmatch(currentLine)
			symbol := res[2]
			date := res[1]
			fmt.Println(symbol)
			fmt.Println(date)

			t, err := time.Parse(timeLayout, res[1])
			if err != nil {
				panic("Error parsing time")
			}

			line2, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Strategy: (.*)")
			strategy := r.FindStringSubmatch(line2)[1]
			fmt.Println(strategy)

			line3, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Constants: \\[([0-9]+\\.[0-9]+)\\s?([0-9]+\\.[0-9]+)?\\s?([0-9]+\\.[0-9]+)?\\]")
			constantsString := r.FindStringSubmatch(line3)

			var constants []float64
			for i := 1; i < len(constantsString); i++ {
				if constantsString[i] != "" {
					num, _ := strconv.ParseFloat(constantsString[i], 64)
					constants = append(constants, num)
				}
			}
			fmt.Println(constants)

			line4, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Sell Price: (.*)")
			price := r.FindStringSubmatch(line4)[1]
			fmt.Println(price)

			// Buy Price: 16.954000

			line4, _ = reader.ReadString('\n')
			r, _ = regexp.Compile("Updated Balance: (.*)€")
			balance, _ := strconv.ParseFloat(r.FindStringSubmatch(line4)[1], 64)
			balance = balance / 10
			fmt.Println(balance)

			line5, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Gain/Loss: (.*)€")
			gain, _ := strconv.ParseFloat(r.FindStringSubmatch(line5)[1], 64)
			fmt.Println(gain)

			line6, _ := reader.ReadString('\n')
			r, _ = regexp.Compile("Profit: (.*)%")
			profit, _ := strconv.ParseFloat(r.FindStringSubmatch(line6)[1], 64)
			fmt.Println(profit)

			if keyExists(symbol, openOrders) {

				order1 := openOrders[symbol]

				order1Quant, _ := strconv.ParseFloat(order1.ExecutedQuantity, 64)
				finalquant := order1Quant - (order1Quant * (1 + profit/100))

				order2 := database.Order{
					Price:                   price,
					Time:                    t.Unix(),
					ExecutedQuantity:        fmt.Sprintf("%f", finalquant),
					CumulativeQuoteQuantity: fmt.Sprintf("%f", finalquant),
					OrigQuantity:            fmt.Sprintf("%f", finalquant),
				}

				var c2 []database.Constant
				for _, constant := range constants {
					c2 = append(c2, database.Constant{Value: constant})
				}
				position := database.Position{
					Symbol:    symbol,
					EntryTime: order1.Time,
					ExitTime:  t.Unix(),
					Orders:    []database.Order{order1, order2},
					Strategy:  strategy,
					Constants: c2,
					Profit:    profit,
					Gain:      finalquant * -1,
				}

				db.Create(&position)

				delete(openOrders, symbol)
			}
		}
	}
}

func keyExists(key string, orders map[string]database.Order) bool {
	if _, ok := orders[key]; ok {
		return true
	}
	return false
}
