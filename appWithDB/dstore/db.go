package dstore

import (
	"github.com/jinzhu/gorm"
)

var globalDBOpt *gorm.DB

func newDBClientPool(name, connectString string) (err error) {
	db, err := gorm.Open(name, connectString)
	if err != nil {
		return
	}

	globalDBOpt = db

	return
}

func migrateTables(tables []interface{}) (err error) {
	for _, one := range tables {
		err = globalDBOpt.AutoMigrate(one).Error
		if err != nil {
			break
		}
	}
	return
}

// LoopExcuteROWs loop Rows selected by sql
func LoopExcuteROWs(cb func(), sql string, pairs ...interface{}) (err error) {
	iter, err := globalDBOpt.Raw(sql).Rows()
	if err != nil {
		return
	}
	defer iter.Close()

	for iter.Next() {
		err = iter.Scan(pairs...)
		if err != nil {
			return
		}
		cb()
	}

	return
}
