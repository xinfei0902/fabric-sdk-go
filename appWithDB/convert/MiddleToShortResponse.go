package convert

import (
	"../fabclient"
	"../tools"
)

// BlockShortResponse struct for Web service
type BlockShortResponse struct {
	Height       uint64 `gorm:"UNIQUE_INDEX;not null" json:"height"`
	PreviousHash string `gorm:"type:varchar(255)" json:"preHash"`
	DataHash     string `gorm:"type:varchar(255)" json:"dataHash"`

	Transactions []TranShortResponse `json:"data,omitempty"`
}

// TranShortResponse record tran information
type TranShortResponse struct {
	ChannelID string `gorm:"type:varchar(32)" json:"channelID,omitempty"`
	TxID      string `gorm:"type:varchar(255);UNIQUE_INDEX;not null" json:"txid,omitempty"`
	Type      string `json:"type,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`

	Transaction []TranDataShortResponse `json:"transaction,omitempty"`

	Config interface{} `json:"config,omitempty"`
}

// TranDataShortResponse record Request / Response data
type TranDataShortResponse struct {
	ChainCodeName    string `gorm:"type:varchar(32)" json:"ccName,omitempty"`
	ChainCodeVersion string `gorm:"type:varchar(32)" json:"ccVersion,omitempty"`

	RespStatus int32  `json:"respStatus,omitempty"`
	RespMsg    string `gorm:"type:text" json:"respMsg,omitempty"`

	ReqType string `gorm:"type:varchar(32)" json:"reqType,omitempty"`
}

// EventBlockToShortResponse convert trans struct into response struct
func EventBlockToShortResponse(middle *fabclient.MiddleCommonBlock) (ret *BlockShortResponse) {
	if middle == nil {
		return
	}
	ret = &BlockShortResponse{
		Height:       middle.Number,
		PreviousHash: tools.Base64Encode(middle.PreviousHash),
		DataHash:     tools.Base64Encode(middle.DataHash),
		Transactions: TransactionsBlockToShortResponse(middle.Child),
	}

	return
}

// TransactionsBlockToShortResponse convert trans struct into response struct
func TransactionsBlockToShortResponse(child []fabclient.MiddleTranNode) (ret []TranShortResponse) {
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

		length := len(one.Config)
		if length > 0 {
			two := one.Config[0]

			hold.Config = &two
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
