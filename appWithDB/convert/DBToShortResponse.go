package convert

import (
	"encoding/json"
)

// EventTableToShortResponse convert trans struct into response struct
func EventTableToShortResponse(tables *EventBlockTable) (ret *BlockShortResponse) {
	if tables == nil {
		return
	}
	ret = &BlockShortResponse{
		Height:       tables.Height,
		PreviousHash: tables.PreviousHash,
		DataHash:     tables.DataHash,
		Transactions: TransactionsTableToShortResponse(tables.Transactions),
	}

	return
}

// TransactionsTableToShortResponse convert trans struct into response struct
func TransactionsTableToShortResponse(child []TranInBlockTable) (ret []TranShortResponse) {
	if len(child) == 0 {
		return nil
	}
	ret = make([]TranShortResponse, 0, len(child))
	for _, one := range child {
		hold := TranShortResponse{}
		hold.ChannelID = one.ChannelID
		hold.TxID = one.TxID
		hold.Type = MiddleTranNodeTypeToString(one.Type)
		hold.Timestamp = one.Timestamp

		length := len(one.ConfigTable)
		if length > 0 {
			two := one.ConfigTable[0]
			buff := []byte(two.Data.RawMessage)

			tmp := make(map[string]interface{})
			err := json.Unmarshal(buff, &tmp)
			if err == nil {
				hold.Config = &tmp
			}
		}

		length = len(one.Transaction)

		if length > 0 {
			hold.Transaction = make([]TranDataShortResponse, length)

			for i, two := range one.Transaction {
				holdTransaction := TranDataShortResponse{}
				holdTransaction.ChainCodeName = two.ChainCodeName
				holdTransaction.ChainCodeVersion = two.ChainCodeVersion
				holdTransaction.RespStatus = two.RespStatus
				holdTransaction.RespMsg = two.RespMsg

				holdTransaction.ReqType = two.ReqType

				hold.Transaction[i] = holdTransaction
			}
		}

		ret = append(ret, hold)
	}
	return
}
