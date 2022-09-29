package database

import (
	"github.com/joho/godotenv"
	"github.com/sdcoffey/techan"
	database "gitlab.com/aoterocom/AOCryptobot/database/models"
	"gitlab.com/aoterocom/AOCryptobot/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"os"
	"time"
)

type DBService struct {
	DB *gorm.DB
}

func NewDBService(dbHost string, dbPort string, dbName string, dbUser string, dbPass string) (*DBService, error) {
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	dbs := &DBService{
		DB: db,
	}

	err = dbs.DB.AutoMigrate(&database.Position{}, &database.Order{}, &database.Candle{}, &database.Constant{})
	if err != nil {
		return nil, err
	}

	return dbs, nil
}

func init() {
	cwd, _ := os.Getwd()
	err := godotenv.Load(cwd + "/conf.env")
	if err != nil {
		log.Fatalln("Error loading go.env file", err)
	}
}

func (dbs *DBService) AddPosition(position models.Position, strategy string, constants []float64,
	profitPct float64, benefits float64, exitTrigger models.ExitTrigger) uint {

	var dbConstants []database.Constant
	for _, constant := range constants {
		dbConstants = append(dbConstants, database.Constant{Value: constant})
	}

	var dbPosition database.Position
	if position.IsClosed() {
		dbPosition = database.Position{
			Symbol:      position.EntranceOrder().Symbol,
			Strategy:    strategy,
			Constants:   dbConstants,
			Profit:      profitPct,
			EntryTime:   position.EntranceOrder().Time,
			ExitTime:    position.ExitOrder().Time,
			Gain:        benefits,
			ExitTrigger: exitTrigger,
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
	} else {
		dbPosition = database.Position{
			Symbol:      position.EntranceOrder().Symbol,
			Strategy:    strategy,
			Constants:   dbConstants,
			Profit:      profitPct,
			EntryTime:   position.EntranceOrder().Time,
			ExitTime:    -1.0,
			Gain:        benefits,
			ExitTrigger: exitTrigger,
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
			}},
		}
	}

	dbs.DB.Create(&dbPosition)
	return dbPosition.ID
}

func (dbs *DBService) GetOpenPositions() []database.Position {
	rows, _ := dbs.DB.Raw("SELECT * FROM positions WHERE profit = -1000").Rows()
	defer rows.Close()

	var positions []database.Position
	for rows.Next() {
		var position database.Position
		dbs.DB.ScanRows(rows, &position)

		rows, _ := dbs.DB.Raw("SELECT * FROM orders WHERE position_id = ? ORDER BY id LIMIT 1", position.ID).Rows()
		var orders []database.Order
		for rows.Next() {
			var order database.Order
			dbs.DB.ScanRows(rows, &order)
			position.Orders = append(position.Orders, order)

			orders = append(orders, order)
			rows.Close()
		}

		rows, _ = dbs.DB.Raw("SELECT * FROM constants WHERE position_id = ?", position.ID).Rows()
		var constants []database.Constant
		for rows.Next() {
			var constant database.Constant
			dbs.DB.ScanRows(rows, &constant)
			constants = append(constants, constant)
		}
		position.Constants = constants

		positions = append(positions, position)
		rows.Close()
	}
	return positions
}

func (dbs *DBService) UpdatePosition(positionID uint, position models.Position, strategy string, constants []float64,
	profitPct float64, benefits float64, exitTrigger models.ExitTrigger) {

	var dbConstants []database.Constant
	for _, constant := range constants {
		dbConstants = append(dbConstants, database.Constant{Value: constant})
	}

	var dbPosition database.Position
	if position.IsClosed() {
		dbPosition = database.Position{
			Symbol:      position.EntranceOrder().Symbol,
			Strategy:    strategy,
			Constants:   dbConstants,
			Profit:      profitPct,
			EntryTime:   position.EntranceOrder().Time,
			ExitTime:    position.ExitOrder().Time,
			Gain:        benefits,
			ExitTrigger: exitTrigger,
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
	} else {
		dbPosition = database.Position{
			Symbol:      position.EntranceOrder().Symbol,
			Strategy:    strategy,
			Constants:   dbConstants,
			Profit:      profitPct,
			EntryTime:   position.EntranceOrder().Time,
			ExitTime:    0,
			Gain:        benefits,
			ExitTrigger: exitTrigger,
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
			}},
		}
	}

	dbs.DB.Unscoped().Delete(&database.Order{}, "position_id = ?", positionID)
	dbs.DB.Unscoped().Delete(&database.Constant{}, "position_id = ?", positionID)
	dbPosition.ID = positionID
	dbPosition.CreatedAt = time.Now()
	dbs.DB.Save(dbPosition)
}

func (dbs *DBService) AddOrUpdateCandle(candle techan.Candle, symbol string) {
	dbCandle := database.Candle{
		Symbol:     symbol,
		Period:     candle.Period.Start.String() + " " + candle.Period.End.String(),
		OpenPrice:  candle.OpenPrice,
		ClosePrice: candle.ClosePrice,
		MaxPrice:   candle.MaxPrice,
		MinPrice:   candle.MinPrice,
		Volume:     candle.Volume,
		TradeCount: candle.TradeCount,
	}

	// Update columns to new value on conflict
	dbs.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}, {Name: "period"}},
		DoUpdates: clause.AssignmentColumns([]string{"symbol", "period", "open_price", "close_price", "max_price", "min_price", "volume", "trade_count"}), // column needed to be updated
	}).Create(&dbCandle)

}

//func (dbs *DBService) databasePositionToOrdersPosition{}
