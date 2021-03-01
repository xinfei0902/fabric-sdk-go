package convert

import (
	"encoding/json"
)

type UserTrace struct {
	Height    uint64 `gorm:"index" json:"height"`
	TxID      string `gorm:"varchar(255);index" json:"txid"`
	Timestamp int64  `gorm:"index" json:"timestamp"`

	Address string `gorm:"varchar(255)" json:"address"`
	Action  string `gorm:"varchar(255)" json:"action"`

	Req        string `gorm:"text" json:"req"`
	Resp       string `gorm:"text" json:"resp"`
	RespStatus int32  `json:"respstatus"`
}

func getUserTables() []interface{} {
	return []interface{}{
		&UserTrace{},
	}
}

type UserAddrss struct {
	Address *string `json:"address"`
}

func MiddleToUserTrace(input *EventBlockTable) (ret []*UserTrace, ok bool) {

	var action, address string
	height := input.Height

	ret = make([]*UserTrace, 0, len(input.Transactions))

	for _, one := range input.Transactions {
		txid := one.TxID
		timestamp := one.Timestamp

		for _, two := range one.Transaction {
			switch two.ChainCodeName {
			case "SIMNI", "ATCNT", "POENE":
			default:
				continue
			}

			action = ""
			address = ""
			data := make(map[string]interface{})

			one := []interface{}{
				&action,
				&UserAddrss{
					Address: &address,
				},
				&data,
			}
			err := json.Unmarshal([]byte(two.ReqArgs), &one)
			if err != nil {
				continue
			}

			if len(action) == 0 || len(address) == 0 {
				continue
			}

			var req []byte
			if len(data) > 0 {
				req, err = json.Marshal(&data)
				if err != nil {
					req = []byte{}
				}
			}

			resp := two.RespPayload
			if two.RespStatus != 200 {
				resp = two.RespMsg
			}

			ret = append(ret, &UserTrace{
				Height:     height,
				TxID:       txid,
				Timestamp:  timestamp,
				Action:     action,
				Address:    address,
				Req:        string(req),
				Resp:       resp,
				RespStatus: two.RespStatus,
			})
		}
	}

	ok = len(ret) > 0

	return
}
