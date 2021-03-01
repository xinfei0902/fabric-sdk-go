package convert

import (
	"../fabclient"
	"../tools"
)

// BlockResponse struct for Web service
type BlockResponse struct {
	SourceURL    string `gorm:"type:varchar(255)" json:"sourceUrl"`
	Height       uint64 `gorm:"UNIQUE_INDEX;not null" json:"height"`
	PreviousHash string `gorm:"type:varchar(255)" json:"preHash"`
	DataHash     string `gorm:"type:varchar(255)" json:"dataHash"`

	Transactions []TranResponse `json:"data,omitempty"`
}

// TranResponse record tran information
type TranResponse struct {
	ChannelID string `gorm:"type:varchar(32)" json:"channelID,omitempty"`
	TxID      string `gorm:"type:varchar(255);UNIQUE_INDEX;not null" json:"txid,omitempty"`
	Type      string `json:"type,omitempty"`
	Epoch     uint64 `json:"epoch,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Version   int32  `json:"version,omitempty"`

	Transaction []TranDataResponse `json:"transaction,omitempty"`

	Config interface{} `json:"config,omitempty"`
}

// TranDataResponse record Request / Response data
type TranDataResponse struct {
	ChainCodeName    string `gorm:"type:varchar(32)" json:"ccName,omitempty"`
	ChainCodeVersion string `gorm:"type:varchar(32)" json:"ccVersion,omitempty"`

	RespStatus  int32  `json:"respStatus,omitempty"`
	RespMsg     string `gorm:"type:text" json:"respMsg,omitempty"`
	RespPayload string `gorm:"type:text" json:"respPayload,omitempty"`

	ReqCCName    string `gorm:"type:varchar(32)" json:"reqccName,omitempty"`
	ReqCCPath    string `gorm:"type:varchar(32)" json:"reqccPath,omitempty"`
	ReqCCVersion string `gorm:"type:varchar(32)" json:"reqccVersion,omitempty"`
	ReqArgs      string `gorm:"type:text"  json:"reqArgs,omitempty"`
	ReqTimeout   int32  `json:"reqTimeout,omitempty"`
	ReqType      string `gorm:"type:varchar(32)" json:"reqType,omitempty"`
}

// EventBlockToResponse convert trans struct into response struct
func EventBlockToResponse(middle *fabclient.MiddleCommonBlock) (ret *BlockResponse) {
	if middle == nil {
		return
	}
	ret = &BlockResponse{
		Height:       middle.Number,
		PreviousHash: tools.Base64Encode(middle.PreviousHash),
		DataHash:     tools.Base64Encode(middle.DataHash),
		SourceURL:    middle.SourceURL,
		Transactions: TransactionsBlockToResponse(middle.Child),
	}

	return
}

// TransactionsBlockToResponse convert trans struct into response struct
func TransactionsBlockToResponse(child []fabclient.MiddleTranNode) (ret []TranResponse) {
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

		length := len(one.Config)
		if length > 0 {
			two := one.Config[0]

			hold.Config = &two
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
				holdTransaction.RespPayload = payloadToString(two.RespPayload)

				holdTransaction.ReqCCName = two.ReqCCName
				holdTransaction.ReqCCPath = two.ReqCCPath
				holdTransaction.ReqCCVersion = two.ReqCCVersion

				holdTransaction.ReqArgs = argsBytesToString(two.ReqArgs)
				holdTransaction.ReqTimeout = two.ReqTimeout
				holdTransaction.ReqType = two.ReqType

				hold.Transaction[i] = holdTransaction
			}
		}

		ret = append(ret, hold)
	}
	return
}
