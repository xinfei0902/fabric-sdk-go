package convert

import "encoding/json"

// EventTableToResponse convert trans struct into response struct
func EventTableToResponse(tables *EventBlockTable) (ret *BlockResponse) {
	if tables == nil {
		return
	}
	ret = &BlockResponse{
		Height:       tables.Height,
		PreviousHash: tables.PreviousHash,
		DataHash:     tables.DataHash,
		SourceURL:    tables.SourceURL,
		Transactions: TransactionsTableToResponse(tables.Transactions),
	}

	return
}

// TransactionsTableToResponse convert trans struct into response struct
func TransactionsTableToResponse(child []TranInBlockTable) (ret []TranResponse) {
	if len(child) == 0 {
		return nil
	}
	ret = make([]TranResponse, 0, len(child))
	for _, one := range child {
		hold := TranResponse{}
		hold.ChannelID = one.ChannelID
		hold.TxID = one.TxID
		hold.Type = MiddleTranNodeTypeToString(one.Type)
		hold.Epoch = one.Epoch
		hold.Timestamp = one.Timestamp
		hold.Version = one.Version

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
			hold.Transaction = make([]TranDataResponse, length)

			for i, two := range one.Transaction {
				holdTransaction := TranDataResponse{}
				holdTransaction.ChainCodeName = two.ChainCodeName
				holdTransaction.ChainCodeVersion = two.ChainCodeVersion
				holdTransaction.RespStatus = two.RespStatus
				holdTransaction.RespMsg = two.RespMsg
				holdTransaction.RespPayload = two.RespPayload

				holdTransaction.ReqCCName = two.ReqCCName
				holdTransaction.ReqCCPath = two.ReqCCPath
				holdTransaction.ReqCCVersion = two.ReqCCVersion

				holdTransaction.ReqArgs = two.ReqArgs
				holdTransaction.ReqTimeout = two.ReqTimeout
				holdTransaction.ReqType = two.ReqType

				hold.Transaction[i] = holdTransaction
			}
		}

		ret = append(ret, hold)
	}
	return
}
