package dstore

import (
	"../convert"
)

// FetchBlock by height from DB
func FetchBlock(height uint64) (ret *convert.EventBlockTable, err error) {
	ret = &convert.EventBlockTable{}
	db := convert.EventBlockTableWhereEqual(globalDBOpt, "Height", height)
	err = db.First(ret).Error
	return
}

// FetchBlockHash by hash from DB
func FetchBlockHash(hash string) (ret *convert.EventBlockTable, err error) {
	ret = &convert.EventBlockTable{}
	db := convert.EventBlockTableWhereEqual(globalDBOpt, "DataHash", hash)
	err = db.First(ret).Error
	return
}

// FetchBlockTxID by txid from DB
func FetchBlockTxID(txid string) (ret *convert.EventBlockTable, err error) {
	ret = &convert.EventBlockTable{}
	value := convert.BuildQueryExpr(globalDBOpt, "TranInBlockTables", "EventBlockTableID", "TxID", txid)
	db := convert.EventBlockTableWhereEqual(globalDBOpt, "ID", value)
	err = db.First(ret).Error
	return
}

// FetchBlockRange by height range from DB
func FetchBlockRange(start, end uint64) (ret []*convert.EventBlockTable, err error) {
	ret = make([]*convert.EventBlockTable, 0, end-start+1)
	db := convert.EventBlockTableBetween(globalDBOpt, "Height", start, end)
	err = db.Find(&ret).Error
	return
}

// FetchTranscations by txid from DB
func FetchTranscations(txid string) (ret *convert.TranInBlockTable, err error) {
	ret = &convert.TranInBlockTable{}
	db := convert.TranInBlockTableEqual(globalDBOpt, "TxID", txid)
	err = db.Find(&ret).Error
	return
}

// FetchUserTrace by address & timestamp
func FetchUserTrace(address string, start, end int64) (ret []*convert.UserTrace, err error) {
	ret = make([]*convert.UserTrace, 0, 16)

	db := convert.UserTraceAddressAndHeight(globalDBOpt, address, start, end)
	err = db.Find(&ret).Error
	return
}
