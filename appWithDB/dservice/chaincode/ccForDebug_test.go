package chaincode

import (
	"fmt"
	"testing"

	"../../dcache"
	"../../djson"
	"github.com/pkg/errors"
)

// TestChangeValue for web.ServiceHandle
func ChangeValue(body []byte) (ret []byte, err error) {
	//Unmarshal request body
	var atcnt Transaction
	err = djson.Unmarshal(body, &atcnt)
	if err != nil {
		err = errors.WithMessage(err, "Parse input body false")
		return nil, err
	}
	atcnt.To[0].Amount = 1000000
	ret, err = djson.Marshal(atcnt)
	if err != nil {
		err = errors.WithMessage(err, "Marshal new value false")
		return nil, err
	}
	return ret, nil
}

// Test_ChangeValue for chaincode userquery
func Test_ChangeValue(t *testing.T) {
	var make UserRegisterMake
	make.path = "/asset/transaction/changeValue"
	make.invoke = true
	make.chaincode = dcache.CCATCNT
	make.function = dcache.MsgForastTestChangeValue
	make.document = ""
	var err error

	body := []byte("{" +
		"\"address\": \"32f8831db3252513600671c92bfea3d9\"," +
		"\"to\": [{" +
		"\"address\": \"00ec55cbe35650e20a964559a339ca24\"," +
		"\"amount\": 500," +
		"\"symbol\": \"GOLD\"" +
		"}" +
		"]," +
		"\"desc\": \"还9月份贷款\"" +
		"}")
	fmt.Println(string(body))
	switch make.function {
	//篡改交易数据，查看交易结果
	case dcache.MsgForastTestChangeValue:
		body, err = ChangeValue(body)
		if err != nil {
			err = errors.WithMessage(err, "change value false.")
			return
		}
		make.function = dcache.MsgForastTrade
	}
	fmt.Println(string(body))

}
