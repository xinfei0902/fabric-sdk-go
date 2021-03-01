package convert

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func init() {
	resigterTables = make([]interface{}, 0, 64)
	resigterTables = append(resigterTables, getChainTables()...)
	resigterTables = append(resigterTables, getPeerSystemTables()...)
	resigterTables = append(resigterTables, getUserTables()...)
}

var resigterTables []interface{}

func getChainTables() []interface{} {
	return []interface{}{
		&EventBlockTable{},
		&TranInBlockTable{},
		&TranActionInBlockTable{},
		&TranCCEventInBlockTable{},
		&TranDataInBlockTable{},
		&ConfigTable{},
	}
}

// GetBlocksTables for migrate Tables
func GetBlocksTables() (ret []interface{}) {
	return resigterTables
}

// BlockAction Keys
const (
	BlockActionWrite = "write"
	BlockActionRead  = "read"
	BlockActionQuery = "query"
)

// ToDBName Convert string into standard format
func ToDBName(input string) string {
	return gorm.ToDBName(input)
}

func equalKey(input string) string {
	return fmt.Sprintf("%s = ?", ToDBName(input))
}

func betweenKey(input string) string {
	input = ToDBName(input)
	return fmt.Sprintf("%s >= ? and %s < ?", input, input)
}

func EventBlockTableBetween(db *gorm.DB, key string, value1, value2 interface{}) (ret *gorm.DB) {
	db = db.Model(&EventBlockTable{}).Where(betweenKey(key), value1, value2)
	ret = db.Preload("Transactions.Transaction").Preload("Transactions.Events").Preload("Transactions.Actions")
	return ret
}

func EventBlockTableWhereEqual(db *gorm.DB, key string, value interface{}) (ret *gorm.DB) {
	db = db.Model(&EventBlockTable{}).Where(equalKey(key), value)
	ret = db.Preload("Transactions.Transaction").Preload("Transactions.Events").Preload("Transactions.Actions").Preload("Transactions.ConfigTable")
	return ret
}

func TranInBlockTableEqual(db *gorm.DB, key string, value interface{}) (ret *gorm.DB) {
	db = db.Model(&TranInBlockTable{}).Where(equalKey(key), value)
	ret = db.Preload("Transaction").Preload("Events").Preload("Actions").Preload("ConfigTable")
	return ret
}

func BuildQueryExpr(db *gorm.DB, table, choice, key string, value interface{}) interface{} {
	table = ToDBName(table)
	choice = ToDBName(choice)
	key = ToDBName(key)
	return db.Table(table).Select(choice).Where(equalKey(key), value).SubQuery()
}

func UserTraceAddressAndHeight(db *gorm.DB, address string, start, end int64) (ret *gorm.DB) {
	db = db.Model(&UserTrace{}).Where("address = ? and height >= ? and height <= ?", address, start, end)
	return db
}
