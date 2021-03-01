package convert

import (
	"encoding/json"
	"fmt"

	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// MiddleTranNodeTypeToString int ==> string
func MiddleTranNodeTypeToString(n int32) string {
	ret, ok := cb.HeaderType_name[n]
	if ok {
		return ret
	}
	return fmt.Sprintf("%v", n)
}

func SafeStringInDB(input string) string {
	buff, err := json.Marshal(input)
	if err != nil {
		return input
	}
	return string(buff)
}
