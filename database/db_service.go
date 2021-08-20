package database

import (
	"github.com/joho/godotenv"
	database "gitlab.com/aoterocom/AOCryptobot/database/models"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

type DBService struct {
	DBName string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
}

func NewDBService(database string, host string, port string, user string, password string) *DBService {
	return &DBService{
		DBName: database,
		DBHost: host,
		DBPort: port,
		DBUser: user,
		DBPass: password,
	}
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/bot_signal-trader/conf.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
}


func (dbs *DBService) AddPosition(position models.Position, strategy string, constants []float64,
	profitPct float64, benefits float64, cumulatedGain float64) {

	var dbConstants []database.Constant
	for _, constant := range constants {
		dbConstants = append(dbConstants, database.Constant{Value: constant})
	}

	dbPosition := database.Position{
		Symbol:        position.EntranceOrder().Symbol,
		Strategy:      strategy,
		Constants:     dbConstants,
		Profit:        profitPct,
		EntryTime:     position.EntranceOrder().Time,
		ExitTime:      position.ExitOrder().Time,
		Gain:          benefits,
		CumulatedGain: cumulatedGain,
		Orders: []database.Order{{
			OrderID:                 position.EntranceOrder().OrderID,
			ClientOrderID:           position.EntranceOrder().ClientOrderID,
			Price:                   position.EntranceOrder().Price,
			OrigQuantity:            position.EntranceOrder().OrigQuantity,
			ExecutedQuantity:        position.EntranceOrder().ExecutedQuantity,
			CumulativeQuoteQuantity: position.EntranceOrder().CumulativeQuoteQuantity,
			Status:                  database.OrderStatusType(position.EntranceOrder().Status),
			Type:                    database.OrderType(position.EntranceOrder().Type),
			Side:                    database.SideType(position.EntranceOrder().Side),
			StopPrice:               position.EntranceOrder().StopPrice,
			IcebergQuantity:         position.EntranceOrder().IcebergQuantity,
			Time:                    position.EntranceOrder().Time,
			UpdateTime:              position.EntranceOrder().UpdateTime,
			IsWorking:               position.EntranceOrder().IsWorking,
			IsIsolated:              position.EntranceOrder().IsIsolated,
		}, {
			OrderID:                 position.ExitOrder().OrderID,
			ClientOrderID:           position.ExitOrder().ClientOrderID,
			Price:                   position.ExitOrder().Price,
			OrigQuantity:            position.ExitOrder().OrigQuantity,
			ExecutedQuantity:        position.ExitOrder().ExecutedQuantity,
			CumulativeQuoteQuantity: position.ExitOrder().CumulativeQuoteQuantity,
			Status:                  database.OrderStatusType(position.ExitOrder().Status),
			Type:                    database.OrderType(position.ExitOrder().Type),
			Side:                    database.SideType(position.ExitOrder().Side),
			StopPrice:               position.ExitOrder().StopPrice,
			IcebergQuantity:         position.ExitOrder().IcebergQuantity,
			Time:                    position.ExitOrder().Time,
			UpdateTime:              position.ExitOrder().UpdateTime,
			IsWorking:               position.ExitOrder().IsWorking,
			IsIsolated:              position.ExitOrder().IsIsolated,
		}},
	}

	dsn := dbs.DBUser + ":" + dbs.DBPass + "@tcp(" + dbs.DBHost + ":" + dbs.DBPort + ")/" + dbs.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.Create(&dbPosition)
}
