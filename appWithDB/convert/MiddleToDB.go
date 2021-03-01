package convert

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"

	"../fabclient"
	"../tools"
)

// ConfigTable record first block information
type ConfigTable struct {
	gorm.Model         `json:"-"`
	TranInBlockTableID uint `gorm:"index" json:"-"`

	Data postgres.Jsonb `json:"data,omitempty"`
}

// EventBlockTable record block information
type EventBlockTable struct {
	gorm.Model `json:"-"`

	SourceURL    string `gorm:"type:varchar(255)" json:"sourceUrl"`
	Height       uint64 `gorm:"UNIQUE_INDEX;not null" json:"height"`
	PreviousHash string `gorm:"type:varchar(255)" json:"preHash"`
	DataHash     string `gorm:"type:varchar(255);index" json:"dataHash"`

	Transactions []TranInBlockTable `json:"data,omitempty"`
}

// TranInBlockTable record tran information
type TranInBlockTable struct {
	gorm.Model        `json:"-"`
	EventBlockTableID uint `gorm:"index" json:"-"`

	// 	TxValidationString string `gorm:"type:varchar(255)"`

	ChannelID string `gorm:"type:varchar(32)" json:"channelID,omitempty"`
	TxID      string `gorm:"type:varchar(255);UNIQUE_INDEX;not null" json:"txid,omitempty"`
	Type      int32  `json:"type,omitempty"`
	Epoch     uint64 `json:"epoch,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Version   int32  `json:"version,omitempty"`

	Transaction []TranDataInBlockTable `json:"transaction,omitempty"`

	Events []TranCCEventInBlockTable `json:"events,omitempty"`

	Actions []TranActionInBlockTable `json:"actions,omitempty"`

	ConfigTable []ConfigTable `json:"config,omitempty"`
}

// TranActionInBlockTable record actions in blocks
type TranActionInBlockTable struct {
	gorm.Model         `json:"-"`
	TranInBlockTableID uint `gorm:"index" json:"-"`

	Namespace string `gorm:"type:varchar(32)" json:"nameSpace,omitempty"`
	Key       string `gorm:"type:varchar(255)" json:"key,omitempty"`
	Value     string `gorm:"type:text" json:"value,omitempty"`
	IsDelete  bool   `json:"isDelete,omitempty"`
	TxNum     uint64 `json:"txNum,omitempty"`
	BlockNum  uint64 `json:"blockNum,omitempty"`
	ReadCount uint64 `json:"readCount,omitempty"`
	EndKey    string `gorm:"type:varchar(255)" json:"endKey,omitempty"`
	Type      string `gorm:"type:varchar(32)" json:"type,omitempty"`
}

// TranCCEventInBlockTable record events
type TranCCEventInBlockTable struct {
	gorm.Model         `json:"-"`
	TranInBlockTableID uint `gorm:"index" json:"-"`

	EventChainCodeID string `gorm:"type:varchar(32)" json:"eventccName,omitempty"`
	EventName        string `gorm:"type:varchar(32)" json:"eventName,omitempty"`
	EventPayload     string `gorm:"type:text" json:"eventPayload,omitempty"`
	EventTxID        string `gorm:"type:varchar(255)" json:"eventTxid,omitempty"`
}

// TranDataInBlockTable record Request / Response data
type TranDataInBlockTable struct {
	gorm.Model         `json:"-"`
	TranInBlockTableID uint `gorm:"index" json:"-"`

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

// EventBlockToTable convert block struct into tables struct
func EventBlockToTable(middle *fabclient.MiddleCommonBlock) (ret *EventBlockTable) {
	if middle == nil {
		return
	}
	ret = &EventBlockTable{
		Height:       middle.Number,
		PreviousHash: tools.Base64Encode(middle.PreviousHash),
		DataHash:     tools.Base64Encode(middle.DataHash),
		SourceURL:    middle.SourceURL,
		Transactions: TransactionsBlockToTable(middle.Child),
	}

	return
}

func ConfigToConfigTable(one *fabclient.ConfigEnvelope) (ret *ConfigTable) {

	buff, err := json.Marshal(one)
	if err != nil || len(buff) == 0 {
		return nil
	}
	metadata := json.RawMessage(buff)
	ret = &ConfigTable{
		Data: postgres.Jsonb{RawMessage: metadata},
	}

	return
}

// TransactionsBlockToTable convert trans struct into tables struct
func TransactionsBlockToTable(child []fabclient.MiddleTranNode) (ret []TranInBlockTable) {
	if len(child) == 0 {
		return nil
	}
	ret = make([]TranInBlockTable, 0, len(child))
	for _, one := range child {
		hold := TranInBlockTable{}
		hold.ChannelID = one.ChannelID
		hold.TxID = one.TxID
		hold.Type = one.Type
		hold.Epoch = one.Epoch
		hold.Timestamp = one.Timestamp
		hold.Version = one.Version

		if len(one.Config) > 0 {
			hold.ConfigTable = make([]ConfigTable, len(one.Config))
			for i, conf := range one.Config {
				tmp := ConfigToConfigTable(&conf)
				if tmp == nil {
					continue
				}
				hold.ConfigTable[i] = *tmp
			}
		}

		length := len(one.Transaction)

		if length > 0 {
			hold.Transaction = make([]TranDataInBlockTable, length)
			hold.Events = make([]TranCCEventInBlockTable, 0, length)
			hold.Actions = make([]TranActionInBlockTable, 0, length)

			for i, two := range one.Transaction {
				holdTransaction := TranDataInBlockTable{}
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

				if len(two.EventTxID) > 0 {
					holdEvent := TranCCEventInBlockTable{}
					holdEvent.EventChainCodeID = two.EventChainCodeID
					holdEvent.EventName = two.EventName
					holdEvent.EventPayload = payloadToString(two.EventPayload)
					holdEvent.EventTxID = two.EventTxID

					hold.Events = append(hold.Events, holdEvent)
				}

				if len(two.Result) > 0 {
					tmp := make([]TranActionInBlockTable, 0, len(two.Result)*3+1)
					for _, three := range two.Result {
						if len(three.Read) > 0 {
							for _, readOne := range three.Read {
								holdAction := TranActionInBlockTable{}

								holdAction.Type = BlockActionRead
								holdAction.Key = SafeStringInDB(readOne.Key)
								holdAction.TxNum = readOne.VersionTxNum
								holdAction.BlockNum = readOne.VersionBlockNum

								tmp = append(tmp, holdAction)
							}
						}
						if len(three.Write) > 0 {
							for _, writeOne := range three.Write {
								holdAction := TranActionInBlockTable{}

								holdAction.Type = BlockActionWrite
								holdAction.Key = SafeStringInDB(writeOne.Key)
								holdAction.Value = payloadToString(writeOne.Value)
								holdAction.IsDelete = writeOne.IsDelete

								tmp = append(tmp, holdAction)
							}
						}
						if len(three.Query) > 0 {
							for _, query := range three.Query {
								holdAction := TranActionInBlockTable{}

								holdAction.Type = BlockActionQuery
								holdAction.Key = SafeStringInDB(query.StartKey)
								holdAction.EndKey = SafeStringInDB(query.EndKey)
								holdAction.ReadCount = uint64(len(query.Reads))

								tmp = append(hold.Actions, holdAction)
							}
						}
					}
					hold.Actions = append(hold.Actions, tmp...)
				}
			}
		}

		ret = append(ret, hold)
	}
	return
}

func payloadToString(input []byte) string {
	if len(input) == 0 {
		return ""
	}
	obj := payloadToObj(input)
	switch obj.(type) {
	case string:
		return obj.(string)
	default:
	}
	buff, err := json.Marshal(&obj)
	if err != nil {
		return tools.Base64Encode(input)
	}
	return string(buff)
}

func payloadToObj(payload []byte) interface{} {
	tmp := bytes.TrimSpace(payload)
	if len(tmp) == 0 {
		if len(payload) == 0 {
			return ""
		}
		return tools.Base64Encode(payload)
	}

	if tmp[0] == '{' && tmp[len(tmp)-1] == '}' {
		obj := make(map[string]interface{})
		err := json.Unmarshal(tmp, &obj)
		if err == nil {
			return obj
		}
	} else if tmp[0] == '[' && tmp[len(tmp)-1] == ']' {
		obj := make([]interface{}, 0, 1)
		err := json.Unmarshal(tmp, &obj)
		if err == nil {
			return obj
		}
	}

	for _, c := range tmp {
		if false == tools.IsAscII(c) {
			return tools.Base64Encode(payload)
		}
	}
	return string(payload)
}

func argsBytesToString(args [][]byte) string {
	if len(args) == 0 {
		return ""
	}
	objList := make([]interface{}, len(args))
	for i, one := range args {
		objList[i] = payloadToObj(one)
	}
	buff, err := json.Marshal(&objList)
	if err == nil {
		return string(buff)
	}

	buff, err = json.Marshal(&args)
	if err == nil {
		return string(buff)
	}

	strList := make([]string, len(args))
	for i, one := range args {
		strList[i] = tools.Base64Encode(one)
	}

	return `["` + strings.Join(strList, `","`) + `"]`
}
